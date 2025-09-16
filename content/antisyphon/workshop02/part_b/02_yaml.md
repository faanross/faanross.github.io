---
showTableOfContents: true
title: "YAML-based Configuration Management System"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson02_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson02_Done).

## Overview
We saw in the previous lesson that we have a `Config` struct in our application acting as the type to house all our different configuration properties.

Now there are many different ways we can use this.

As we already saw, we can just define a struct literal right there in our `main` function and assign the values to it. This is the most direct, bare bones technique.

Another option would be to create a constructor which we could house in our `config` package, and then call it from `main` by passing all the desired values as arguments. This is on one hand a bit more contrived, but it keeps our `main` function clean and also is more conducive to the implementation of validation logic.

A way I prefer to typically handle configs however is by implementing a YAML-based system. I like this for the simple reason that I believe it creates a more user-friendly interface when we specify our desired config values. Instead of looking through our code where we're specifying the values, we have a clean, separate file written in YAML which is probably the closest you're gonna get to pure English in development.

Now, as was the case in the previous lesson when I showed how to create an embedded struct - in all honesty, creating a YAML-based implementation system is probably a little bit overkill for this application. This is because it comes with overhead - we now also have to implement a loader, which will read the YAML, create the struct, and unmarshall the YAML values into the struct.

I'm once again choosing to do this since I think it's a great touch when projects become larger and more complex, and so I wanted you to know how you can do this. Besides, it's really not that much effort, and as I just said, in larger projects this extra step is gonna pay off in terms of an improved user experience.



## What We'll Create
- YAML config (`./configs/config.yaml`)
- YAML config loader (`internals/config/loader.go`)


## Import Library
To work with YAML in go we need to import the following package:

```bash
go get gopkg.in/yaml.v3
```



## Adding tags to our Config struct

Before we create a YAML file for our config values, we need to add YAML-tags to our existing struct. This is to create a connection between the tag/label of the field in our YAML file, and the field in the struct. Essentially, think of it this way, when we add `yaml:"client"` to the `ClientAddr` field we are saying: This same field is called `client` when its in YAML-format.

By creating these connections we ensure the correct values are unmarshalled from the YAML to our struct.

```go
// Config holds all application configuration
type Config struct {
	ClientAddr string       `yaml:"client"`
	ServerAddr string       `yaml:"server"`
	Timing     TimingConfig `yaml:"timing"`
	Protocol   string       `yaml:"protocol"` // this will be the starting protocol
	TlsKey     string       `yaml:"tls_key"`
	TlsCert    string       `yaml:"tls_cert"`
}

type TimingConfig struct {
	Delay  time.Duration `yaml:"delay"`  // Base delay between cycles
	Jitter int           `yaml:"jitter"` // Jitter percentage (0-100)}
}
```


Great, so now we can create our YAML file.


## config yaml file

Let's add the following to a new file `./configs/config.yaml`. It's important to take note that the labels used here **have to** match the tags we just added to our struct.

```yaml
client: "127.0.0.1:0"

server: "127.0.0.1:8443"

timing:
  delay: "5s"
  jitter: 50

protocol: "https"

tls_key: "./certs/server.key"
tls_cert: "./certs/server.crt"
```

We are referencing `tls_key` and `tls_cert` here, neither of which exists yet. This is fine, we'll add it in our next lesson.

Great, so now we have our config as a simple, clean YAML file, and our struct has the ability, or "knowledge", to receive values that are unmarshalled from a YAML file. What we need now is a function that, after starting our application, will do just that - read the YAML file, create a new struct instance, and then unmarshall the values from the YAML file into the struct.



## Config loader
Let's create this new file in `internals/config/loader.go`.

Let's first just create the function with placeholder comments:

```go
// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {

	// Load YAML file from disk

	// Create empty Config struct

	// Unmarshall YAML file values into struct

	// Optional but good: Validate values

}
```


First thing is we want to load the YAML file from disk, we get this path as the argument called `path`:

```go
	// We'll provide path to *.yaml to function when we call it
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()
```


Now we can create our empty struct, and unmarshall the YAML values into it:

```go
	// instantiate struct to unmarshall yaml into
	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
```


Finally, we can call a yet-to-be created method on `cfg`, which will perform some basic validation:

```go
	// Optional, but good proactive -> Validate the configuration
	if err := cfg.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
```


So our entire function, including the return statement at the end, becomes:

```go

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	
	// We'll provide path to *.yaml to function when we call it
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	// instantiate struct to unmarshall yaml into
	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Optional, but good proactive -> Validate the configuration
	if err := cfg.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
```



## Basic validation
In the same file let's now implement the `ValidateConfig()` method we just called inside of `LoadConfig()`:

```go
// ValidateConfig checks if the configuration is valid
func (c *Config) ValidateConfig() error {
	if c.ClientAddr == "" {
		return fmt.Errorf("agent address cannot be empty")
	}

	if c.ServerAddr == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	if c.Timing.Delay <= 0 {
		return fmt.Errorf("delay must be positive")
	}

	if c.Timing.Jitter < 0 || c.Timing.Jitter > 100 {
		return fmt.Errorf("jitter must be between 0 and 100")
	}

	if c.TlsCert == "" {
		return fmt.Errorf("tls cert cannot be empty")
	}

	if c.TlsKey == "" {
		return fmt.Errorf("tls key cannot be empty")
	}

	return nil
}
```

As you can see it's all extremely basic, we essentially just ensure for most values that there is SOME value, for delay we ensure it is not 0 or negative, and for jitter we ensure it's any value between 0 and 100 inclusive. Feel free to add more robust validation here if you'd like - for example to ensure the provided IPs are valid etc.


## Agent's main.go

I'll add some new contrived logic here so we can test that our set up works: That the values stated in the YAML file are loaded and unmarshalled into a Config struct.


```go

const pathToYAML = "./configs/config.yaml"

func main() {
	// Command line flag for config file path
	configPath := flag.String("config", pathToYAML, "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded configuration:\n")
	log.Printf("-> Client: %s\n", cfg.ClientAddr)
	log.Printf("-> Server: %s\n", cfg.ServerAddr)
	log.Printf("-> Delay: %v\n", cfg.Timing.Delay)
	log.Printf("-> Jitter: %d%%\n", cfg.Timing.Jitter)
	log.Printf("-> Starting Protocol: %s\n", cfg.Protocol)

}
```


Note that I've added add the ability to state the path to the config as a command-line argument using `-config`, but if we omit it, it'll default to `pathToYAML` (`./configs/config.yaml`).


## Test

We thus expect that if we now run our application, the same values we provided in our YAML file should be printed to console.

```shell
❯ go run ./cmd/agent
2025/08/09 12:59:52 Loaded configuration:
2025/08/09 12:59:52 -> Client: 127.0.0.1:0
2025/08/09 12:59:52 -> Server: 127.0.0.1:8443
2025/08/09 12:59:52 -> Delay: 5s
2025/08/09 12:59:52 -> Jitter: 50%
2025/08/09 12:59:52 -> Starting Protocol: https
```


And this is exactly what we see.


## Conclusion

That's it as far as our foundation goes. We have our interfaces, our factory functions, and now we have a config system with a struct, YAML-file, and ability to load values from the latter to populate the former.

We'll now move into our next phase where we'll implement the ability to communicate over HTTPS.







___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "01_interfaces.md" >}})
[|NEXT|]({{< ref "../part_c/01_https_server.md" >}})