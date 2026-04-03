package config

import "errors"

type Profile struct {
	Name            string `json:"name"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	EndpointURL     string `json:"endpoint_url,omitempty"`
	Bucket          string `json:"bucket,omitempty"`
}

func (p *Profile) Validate() error {
	if p.Name == "" {
		return errors.New("profile name is required")
	}
	if p.Region == "" {
		return errors.New("region is required")
	}
	if p.AccessKeyID == "" {
		return errors.New("access_key_id is required")
	}
	if p.SecretAccessKey == "" {
		return errors.New("secret_access_key is required")
	}
	return nil
}

func AddProfile(cfg *Config, profile Profile) error {
	if err := profile.Validate(); err != nil {
		return err
	}

	for _, existing := range cfg.Profiles {
		if existing.Name == profile.Name {
			return errors.New("profile already exists")
		}
	}

	cfg.Profiles = append(cfg.Profiles, profile)
	cfg.ActiveProfile = profile.Name
	return nil
}

func RemoveProfile(cfg *Config, name string) error {
	for i, p := range cfg.Profiles {
		if p.Name == name {
			cfg.Profiles = append(cfg.Profiles[:i], cfg.Profiles[i+1:]...)
			if cfg.ActiveProfile == name {
				cfg.ActiveProfile = ""
			}
			return nil
		}
	}
	return errors.New("profile not found")
}

func SwitchProfile(cfg *Config, name string) error {
	for _, p := range cfg.Profiles {
		if p.Name == name {
			cfg.ActiveProfile = name
			return nil
		}
	}
	return errors.New("profile not found")
}

func UpdateProfile(cfg *Config, name string, updates Profile) error {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			// Update only fields that are provided (non-empty)
			if updates.Region != "" {
				cfg.Profiles[i].Region = updates.Region
			}
			if updates.AccessKeyID != "" {
				cfg.Profiles[i].AccessKeyID = updates.AccessKeyID
			}
			if updates.SecretAccessKey != "" {
				cfg.Profiles[i].SecretAccessKey = updates.SecretAccessKey
			}
			if updates.EndpointURL != "" {
				cfg.Profiles[i].EndpointURL = updates.EndpointURL
			}
			if updates.Bucket != "" {
				cfg.Profiles[i].Bucket = updates.Bucket
			}
			return nil
		}
	}
	return errors.New("profile not found")
}
