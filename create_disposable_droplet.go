// Copyright 2020 DigitalOcean
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"github.com/do-community/do-disposable/copyserver"
	"github.com/buger/goterm"
	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

// Generic error for interrupt.
type interruptType struct{}
func (interruptType) Error() string { return "" }
var interrupt = &interruptType{}

// Check if the connection is IPv6.
func checkIfIpv6() bool {
	reqError := func() bool {
		println("Unable to determine if the connection is IPv6. Going to attempt to default to connecting to the droplet via IPv4.")
		return false
	}
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil || resp.StatusCode != 200 {
		return reqError()
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return reqError()
	}
	matched, err := regexp.Match("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+", b)
	if err != nil {
		return reqError()
	}
	return !matched
}

// This function is used to create the disposable droplet/kill it.
func handleDisposableDroplet(region, size, distro string) {
	// Determine if we are on a IPv6 connection.
	ipv6 := checkIfIpv6()

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

	// Keep trying to connect via SSH until it works.
	var ip string
	network := "tcp"
	if ipv6 {
		ip, _ = d.PublicIPv6()
		network = "tcp6"
	} else {
		ip, _ = d.PublicIPv4()
	}
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
			client, err = ssh.Dial(network, ip+":22", &ssh.ClientConfig{
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

		// Handle copyback/copyfrom init.
		session, err := client.NewSession()
		if err != nil {
			errorChan <- err
			return
		}
		err = session.Run("wget -O - -o /dev/null https://community-tools.sfo2.digitaloceanspaces.com/droplet_init.sh | bash")
		if err != nil {
			errorChan <- err
			return
		}

		// Create a new session.
		session, err = client.NewSession()
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

		// Create a fasthttp server for the SSH server to be able to use.
		go func() {
			// Create the listener on the SSH side.
			listener, err := client.Listen("tcp", "127.0.0.1:8190")
			if err != nil {
				errorChan <- err
				return
			}

			// Create the server.
			err = copyserver.Copyserver(listener)
			if err != nil {
				errorChan <- err
			}
		}()

		// Request pseudo terminal.
		// TODO: The terminal doesn't handle some formatting right.
		err = session.RequestPty("vt100", goterm.Height(), goterm.Width(), modes)
		if err != nil {
			errorChan <- err
			return
		}

		// Loop handling the width/height.
		go func() {
			currentHeight := goterm.Height()
			currentWidth := goterm.Width()
			for dropletActionsActive {
				// Get the width/height.
				w := goterm.Width()
				h := goterm.Height()

				// Check if it's different.
				if w != currentWidth || h != currentHeight {
					// Set the new width/height.
					currentWidth = w
					currentHeight = h
					err := session.WindowChange(h, w)
					if err != nil {
						errorChan <- err
						return
					}
				}

				// Sleep for 100ms.
				time.Sleep(time.Millisecond * 100)
			}
		}()

		// Start SSH shell.
		err = session.Shell()
		if err != nil {
			errorChan <- err
			return
		}

		// Handle input.
		go func() {
			ob := make([]byte, 1)
			for dropletActionsActive {
				_, err := os.Stdin.Read(ob)
				if err != nil {
					errorChan <- err
					return
				}
				if ob[0] == '\r' {
					// Windows.
					continue
				}
				_, err = stdin.Write(ob)
				if err != nil {
					errorChan <- err
					return
				}
			}
		}()

		// Handle waiting for disconnect.
		go func() {
			// Wait for the session.
			err := session.Wait()

			// If this is a exit error, return nil since we don't care about old command errors.
			if _, ok := err.(*ssh.ExitError); ok {
				errorChan <- nil
			}

			// If not, return the error.
			errorChan <- err
		}()
	}()

	// Handle any new errors/the client closing.
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
