package filesystem

import "net/url"

type config interface {
	// Driver returns the driver name.
	Driver() string
	// Host returns the hostname.
	Host() string
	// Path returns the root path.
	Path() string

	// UserInfo returns the user information.
	UserInfo() *url.Userinfo

	// Encode the config to url.Values
	Encode() url.Values

	// Decode the config from url.Values
	Decode(query url.Values) error

	// StringEncode encodes the config to string
	StringEncode() string

	// StringDecode decodes the config from string
	StringDecode(querystring string) error
}

type Config struct {
}

func (conf *Config) Driver() string {
	panic("implement config.Driver()")
}

func (conf *Config) Host() string {
	return ""
}

func (conf *Config) Path() string {
	return ""
}

func (conf *Config) UserInfo() *url.Userinfo {
	return nil
}

func (conf *Config) Encode() url.Values {
	return url.Values{}
}

func (conf *Config) Decode(query url.Values) error {

	return nil
}

func (conf *Config) StringEncode() string {
	return conf.Encode().Encode()
}

func (conf *Config) StringDecode(querystring string) error {
	v, err := url.ParseQuery(querystring)
	if err != nil {
		return err
	}

	return conf.Decode(v)
}

// BuildDsn builds the DSN string from the config.
func BuildDsn(conf config) string {
	return (&url.URL{
		Host:     conf.Host(),
		Path:     conf.Path(),
		User:     conf.UserInfo(),
		Scheme:   conf.Driver(),
		RawQuery: conf.StringEncode(),
	}).String()
}
