package config

import (
	"encoding/json"
	"flag"
	"fmt"
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
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (c *CLI) addProfile(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	name := fs.String("name", "", "Profile name")
	region := fs.String("region", "", "S3 region")
	accessKey := fs.String("access-key", "", "Access key ID")
	secretKey := fs.String("secret", "", "Secret access key")
	endpoint := fs.String("endpoint", "", "Custom endpoint URL (optional)")
	bucket := fs.String("bucket", "", "S3 bucket name (optional)")

	if err := fs.Parse(args); err != nil {
		return err
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
	fs := flag.NewFlagSet("switch", flag.ExitOnError)
	name := fs.String("name", "", "Profile name")

	if err := fs.Parse(args); err != nil {
		return err
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
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	name := fs.String("name", "", "Profile name")

	if err := fs.Parse(args); err != nil {
		return err
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
