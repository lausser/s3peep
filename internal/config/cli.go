package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type CLI struct {
	configPath string
}

func NewCLI(configPath string) *CLI {
	return &CLI{configPath: configPath}
}

func (c *CLI) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: s3peep profile <command>")
	}

	cfg, err := Load(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch args[0] {
	case "add":
		return c.addProfile(cfg, args[1:])
	case "list":
		return c.listProfiles(cfg)
	case "switch":
		return c.switchProfile(cfg, args[1:])
	case "remove":
		return c.removeProfile(cfg, args[1:])
	case "update":
		return c.updateProfile(cfg, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func normalizeFlagError(err error) string {
	errMsg := err.Error()
	// Replace single-dash with double-dash in error messages
	re := regexp.MustCompile(`([:\s"])-(\w+)`)
	errMsg = re.ReplaceAllString(errMsg, "${1}--${2}")
	return errMsg
}

func (c *CLI) addProfile(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	// Override usage to show double dashes - needs to use stderr for error messages
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of add:\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	name := fs.String("name", "", "Profile name")
	region := fs.String("region", "", "S3 region")
	accessKey := fs.String("access-key", "", "Access key ID")
	secretKey := fs.String("secret", "", "Secret access key")
	endpoint := fs.String("endpoint", "", "Custom endpoint URL (optional)")
	bucket := fs.String("bucket", "", "S3 bucket name (optional)")

	if err := fs.Parse(args); err != nil {
		// Normalize error and check for help request
		errMsg := normalizeFlagError(err)
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			fs.Usage()
		}
		// For help request, Usage was already called by flag package
		return fmt.Errorf("")
	}

	if *name == "" || *region == "" || *accessKey == "" || *secretKey == "" {
		return fmt.Errorf("required flags: --name, --region, --access-key, --secret")
	}

	profile := Profile{
		Name:            *name,
		Region:          *region,
		AccessKeyID:     *accessKey,
		SecretAccessKey: *secretKey,
		EndpointURL:     *endpoint,
		Bucket:          *bucket,
	}

	if err := AddProfile(cfg, profile); err != nil {
		return err
	}

	if err := Save(cfg, c.configPath); err != nil {
		return err
	}

	fmt.Printf("Profile '%s' added successfully\n", *name)
	return nil
}

func (c *CLI) listProfiles(cfg *Config) error {
	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured")
		return nil
	}

	fmt.Println("Profiles:")
	for _, p := range cfg.Profiles {
		active := ""
		if p.Name == cfg.ActiveProfile {
			active = " (active)"
		}
		fmt.Printf("  - %s%s\n", p.Name, active)
		if p.EndpointURL != "" {
			fmt.Printf("    Endpoint: %s\n", p.EndpointURL)
		}
		fmt.Printf("    Region: %s, Bucket: %s\n", p.Region, p.Bucket)
	}
	return nil
}

func (c *CLI) switchProfile(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("switch", flag.ContinueOnError)
	// Override usage to show double dashes
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of switch:\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	name := fs.String("name", "", "Profile name")

	if err := fs.Parse(args); err != nil {
		// Normalize error and check for help request
		errMsg := normalizeFlagError(err)
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			fs.Usage()
		}
		// For help request, Usage was already called by flag package
		return fmt.Errorf("")
	}

	if *name == "" {
		return fmt.Errorf("profile name is required")
	}

	if err := SwitchProfile(cfg, *name); err != nil {
		return err
	}

	if err := Save(cfg, c.configPath); err != nil {
		return err
	}

	fmt.Printf("Switched to profile '%s'\n", *name)
	return nil
}

func (c *CLI) removeProfile(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	// Override usage to show double dashes
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of remove:\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	name := fs.String("name", "", "Profile name")

	if err := fs.Parse(args); err != nil {
		// Normalize error and check for help request
		errMsg := normalizeFlagError(err)
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			fs.Usage()
		}
		// For help request, Usage was already called by flag package
		return fmt.Errorf("")
	}

	if *name == "" {
		return fmt.Errorf("profile name is required")
	}

	if err := RemoveProfile(cfg, *name); err != nil {
		return err
	}

	if err := Save(cfg, c.configPath); err != nil {
		return err
	}

	fmt.Printf("Profile '%s' removed\n", *name)
	return nil
}

func (c *CLI) updateProfile(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	// Override usage to show double dashes
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of update:\n")
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s %s\n", f.Name, f.DefValue)
			fmt.Fprintf(os.Stderr, "\t%s\n", f.Usage)
		})
	}
	name := fs.String("name", "", "Profile name (required)")
	region := fs.String("region", "", "S3 region")
	accessKey := fs.String("access-key", "", "Access key ID")
	secretKey := fs.String("secret", "", "Secret access key")
	endpoint := fs.String("endpoint", "", "Custom endpoint URL")
	bucket := fs.String("bucket", "", "S3 bucket name")

	if err := fs.Parse(args); err != nil {
		// Normalize error and check for help request
		errMsg := normalizeFlagError(err)
		if !strings.Contains(errMsg, "help requested") {
			fmt.Fprintf(os.Stderr, "%s\n", errMsg)
			fs.Usage()
		}
		// For help request, Usage was already called by flag package
		return fmt.Errorf("")
	}

	if *name == "" {
		return fmt.Errorf("profile name is required (--name)")
	}

	// Build update profile with provided values (empty strings mean no change)
	updates := Profile{
		Region:          *region,
		AccessKeyID:     *accessKey,
		SecretAccessKey: *secretKey,
		EndpointURL:     *endpoint,
		Bucket:          *bucket,
	}

	if err := UpdateProfile(cfg, *name, updates); err != nil {
		return err
	}

	if err := Save(cfg, c.configPath); err != nil {
		return err
	}

	fmt.Printf("Profile '%s' updated successfully\n", *name)
	return nil
}

func PrintConfigExample() {
	example := Config{
		ActiveProfile: "my-profile",
		Profiles: []Profile{
			{
				Name:            "my-profile",
				Region:          "us-east-1",
				AccessKeyID:     "YOUR_ACCESS_KEY",
				SecretAccessKey: "YOUR_SECRET_KEY",
				Bucket:          "",
				EndpointURL:     "https://s3.amazonaws.com",
			},
		},
	}
	data, _ := json.MarshalIndent(example, "", "  ")
	fmt.Println("Example config:")
	fmt.Println(string(data))
}

func CreateDefaultConfig(path string) error {
	cfg := &Config{
		ActiveProfile: "",
		Profiles:      []Profile{},
	}
	return Save(cfg, path)
}
