package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"strings"
)

// Used to define the DO client.
var client *godo.Client

// Used to get the user to input their token.
func setToken() string {
	for {
		print("What is your DigitalOcean token? You will need a read/write API key which you can generate from the \"API Keys\" panel when you are signed in with your DigitalOcean account: ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text == "" {
			continue
		}
		client = godo.NewFromToken(text)
		_, resp, err := client.Tags.List(context(), &godo.ListOptions{})
		if err == nil {
			return text
		}
		if resp != nil && resp.StatusCode == 401 {
			continue
		}
		panic(err)
	}
}

// Tries to load the config. Returns false if requires init.
func loadConfig() (string, bool) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		// For every platform we support, we expect a home folder to exist.
		// There's not much we can do here except crash.
		panic(err)
	}
	fp := path.Join(homedir, ".do-disposable")
	f, err := os.Open(fp)

	// Handle file being existent/readable.
	if err == nil {
		// Attempt to decode the Gob file.
		decoder := gob.NewDecoder(f)
		err = decoder.Decode(&config)
		if err != nil {
			// The user has either manually tried to modify this or something went wrong in the application.
			panic(err)
		}

		// Create the client with this token.
		client = godo.NewFromToken(config.Token)

		// Return true here.
		return fp, true
	}

	if os.IsNotExist(err) {
		// The file doesn't exist. Return false.
		return fp, false
	} else {
		// The error here is a configuration issue with the users system. We will crash.
		panic(err)
	}
}

// Quick hack to get the tag from a description.
func getTag(x string) string {
	l := strings.Split(x, " ")
	return l[len(l)-1][1:len(l[len(l)-1])-1]
}

// Used to write the config.
func writeConfig(fp string) {
	f, err := os.Create(fp)
	if err != nil {
		panic(err)
	}
	e := gob.NewEncoder(f)
	err = e.Encode(config)
	if err != nil {
		panic(err)
	}
}

// Used to set the default region.
func setDefaultRegion() []godo.Region {
	regions, _, err := client.Regions.List(context(), &godo.ListOptions{})
	if err != nil {
		// Hmmmmm this is odd.
		panic(err)
	}
	descs := make([]string, len(regions))
	nyc3Exists := false
	for i, v := range regions {
		if !v.Available {
			// Not relevant to us. Continue.
			continue
		}
		descs[i] = v.Name + " [" + v.Slug + "]"
		if v.Slug == "nyc3" {
			nyc3Exists = true
		}
	}
	DefaultRegion := "New York 3 [nyc3]"
	if !nyc3Exists {
		DefaultRegion = descs[0]
	}
	config.DefaultRegion = getTag(FormatList("What's the default region you wish to use?", descs, &DefaultRegion))
	return regions
}

// Used to set the droplet size.
func setSize(regions []godo.Region) {
	var err error
	if regions == nil {
		regions, _, err = client.Regions.List(context(), &godo.ListOptions{})
		if err != nil {
			// Hmmmmm this is odd.
			panic(err)
		}
	}
	var sizes []string
	for _, v := range regions {
		if v.Slug == config.DefaultRegion {
			sizes = v.Sizes
		}
	}
	doSizes := map[string]godo.Size{}
	doGlobalSizes, _, err := client.Sizes.List(context(), &godo.ListOptions{})
	if err != nil {
		// Very odd.
		panic(err)
	}
	for _, v := range doGlobalSizes {
		if !v.Available {
			// Ignore this.
			continue
		}
		doSizes[v.Slug] = v
	}
	dropletDescs := make([]string, 0, len(sizes))
	for _, v := range sizes {
		x, ok := doSizes[v]
		if !ok {
			continue
		}
		dropletDescs = append(dropletDescs, fmt.Sprintf("%d GB storage/%d vCPUS/%d MB RAM/$%f per hour [%s]", x.Disk, x.Vcpus, x.Memory, x.PriceHourly, v))
	}
	config.DefaultSize = getTag(FormatList("What's the default droplet size you wish to use?", dropletDescs, &dropletDescs[0]))
}

// Used to get the public SSH key of the user.
func getPublicKey() string {
	pub, err := ssh.NewPublicKey(&config.PrivateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	return string(ssh.MarshalAuthorizedKey(pub))
}

// Used to generate/upload the SSH key.
func genSSHKey() {
	print("Generating application specific SSH key... ")
	var err error
	config.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}
	err = config.PrivateKey.Validate()
	if err != nil {
		panic(err)
	}
	println("done!")

	// Save the SSH key to the user.
	print("Saving application specific SSH key to user... ")
	info, _, err := client.Keys.Create(context(), &godo.KeyCreateRequest{
		Name:      "do-disposable ["+uuid.New().String()+"]",
		PublicKey: getPublicKey(),
	})
	if err != nil {
		panic(err)
	}
	config.KeyID = info.ID
	println("done!")
}

// Used to get the user to input their config options and then save it as a new config.
func inputSaveConfig(fp string) {
	// Create the base config structure.
	config = &configStructure{}

	// Set the users token.
	config.Token = setToken()

	// Create the client with this token.
	client = godo.NewFromToken(config.Token)

	// Get the default region from the user.
	regions := setDefaultRegion()

	// Get the default droplet size from the user.
	setSize(regions)

	// Generate the SSH key.
	genSSHKey()
	
	// Write the config.
	writeConfig(fp)
}

// Used to initialise the DigitalOcean client and config.
func clientInit() {
	_, exists := loadConfig()
	if !exists {
		println("Configuration is not set. Please run do-disposable auth.")
		os.Exit(1)
	}
}
