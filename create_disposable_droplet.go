package main

import (
	"bufio"
	"fmt"
	"github.com/buger/goterm"
	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Generic error for interrupt.
type interruptType struct{}
func (interruptType) Error() string { return "" }
var interrupt = &interruptType{}

// This function is used to create the disposable droplet/kill it.
func handleDisposableDroplet(region, size, distro string) {
	// Defines the droplet ID.
	ID := uuid.New().String()

	// Create the droplet.
	print("Creating droplet... ")
	d, _, err := client.Droplets.Create(context(), &godo.DropletCreateRequest{
		Name:              ID,
		Region:            region,
		Size:              size,
		Image:             godo.DropletCreateImage{Slug: distro},
		SSHKeys:           []godo.DropletCreateSSHKey{{ID: config.KeyID}},
		IPv6:              true,
		Tags:              []string{"do-disposable"},
	})
	if err != nil {
		panic(err)
	}
	println("done!")

	// From here, we should try and ensure that any panic/exit destroys this droplet.
	// The destruction should allow for bad internet connections and should be patient.
	defer func() {
		// Handle describing what happened to the user.
		r := recover()
		if r == nil {
			println("The application was exited. Destroying the droplet before quitting. Note that closing the process before this is done will mean you'll have to manually delete the droplet.")
		} else {
			log.Print(r)
			println("The application crashed. Destroying the droplet before quitting. Note that closing the process before this is done will mean you'll have to manually delete the droplet.")
		}

		// Try destroying the droplet.
		for {
			resp, err := client.Droplets.Delete(context(), d.ID)
			if resp != nil {
				if resp.StatusCode == 404 {
					println("Droplet no longer exists.")
				} else if resp.StatusCode == 401 {
					println("The authorization token was revoked. You will need to manually destroy this droplet!")
				}
			}
			if err == nil {
				break
			}
			log.Println("Failed to delete droplet. Will try again: ", err)
		}

		// Log that the droplet was deleted.
		println("Droplet deleted.")
		if r != nil {
			os.Exit(1)
		}
	}()

	// The channel which is used for errors (and nils to represent moving along).
	errorChan := make(chan error)

	// Defines if the droplet actions should still be running.
	// During a CTRL+C this would be set to false.
	dropletActionsActive := true

	// Handle CTRL+C.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		dropletActionsActive = false
		errorChan <- interrupt
	}()

	// Handle the waiting for the droplet.
	print("Waiting for droplet to be active... ")
	go func() {
		for dropletActionsActive {
			time.Sleep(time.Second)
			d, _, err = client.Droplets.Get(context(), d.ID)
			if err != nil {
				errorChan <- err
				return
			}
			if d.Status == "active" {
				errorChan <- nil
				println("done!")
				return
			}
		}
	}()

	// Handle errors/pass through for the initial creation.
	err = <-errorChan
	if err == interrupt {
		return
	} else if err != nil {
		panic(err)
	}

	// TODO: Handle IPv6 only connections.
	// Keep trying to connect via SSH until it works.
	ipv4, _ := d.PublicIPv4()
	print("Waiting for the droplet to accept SSH connections... ")
	signer, err := ssh.NewSignerFromKey(config.PrivateKey)
	if err != nil {
		panic(err)
	}
	var client *ssh.Client
	var stdin io.WriteCloser
	go func() {
		// Wait for SSH to be ready.
		for {
			client, err = ssh.Dial("tcp", ipv4+":22", &ssh.ClientConfig{
				Config:            ssh.Config{},
				User:              "root",
				Auth:              []ssh.AuthMethod{ssh.PublicKeys(signer)},

				// Sadly it is impossible to make the initial handshake secure here since this is the first time we see the host.
				// After the initial handshake, the rest of the data will be encrypted.
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			if err != nil {
				continue
			}
			break
		}
		println("done!")

		// Get a session.
		session, err := client.NewSession()
		if err != nil {
			errorChan <- err
			return
		}

		// Get the IO pipes.
		session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
		session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
		stdin, _ = session.StdinPipe()

		// Set the terminal modes which we want.
		modes := ssh.TerminalModes{
			ssh.ECHO:  0,
			ssh.IGNCR: 1,
		}

		// Request pseudo terminal.
		// TODO: The terminal doesn't seem to pass in raw key events.
		err = session.RequestPty("xterm", goterm.Height(), goterm.Width(), modes)
		if err != nil {
			errorChan <- err
			return
		}

		// Start SSH shell.
		err = session.Shell()
		if err != nil {
			errorChan <- err
			return
		}

		// Handle input.
		go func() {
			for dropletActionsActive {
				reader := bufio.NewReader(os.Stdin)
				b, err := reader.ReadString('\n')
				if err != nil {
					errorChan <- err
					return
				}
				_, err = fmt.Fprint(stdin, b)
				if err != nil {
					errorChan <- err
					return
				}
			}
		}()
	}()

	// Handle any new errors/the client closing.
	// TODO: Handle connection being killed by remote server.
	for {
		err := <-errorChan
		if err == interrupt {
			if client == nil {
				dropletActionsActive = false
				return
			} else {
				_, err = fmt.Fprint(stdin, "\x03")
				if err != nil {
					panic(err)
				}
			}
		} else if err != nil {
			panic(err)
		} else {
			break
		}
	}
}
