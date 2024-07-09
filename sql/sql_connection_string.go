package sql

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cjlapao/common-go/guard"
)

type SqlConnectionString struct {
	Username  string
	Password  string
	Port      int
	Server    string
	Database  string
	EnableTLS bool
}

func (c *SqlConnectionString) ConnectionString() string {
	if c.Port > 0 {
		c.Server = c.Server + ":" + strconv.Itoa(c.Port)
	}
	result := c.Username + ":" + c.Password + "@tcp(" + c.Server + ")/" + c.Database + "?parseTime=true"
	if c.EnableTLS {
		result += "&tls=true"
	}

	return result
}

func (c *SqlConnectionString) WithUser(username string) *SqlConnectionString {
	c.Username = username
	return c
}

func (c *SqlConnectionString) WithPassword(password string) *SqlConnectionString {
	c.Password = password
	return c
}

func (c *SqlConnectionString) WithServer(serverName string) *SqlConnectionString {
	if strings.ContainsAny(serverName, ":") {
		parts := strings.Split(serverName, ":")
		c.Server = parts[0]
		if port, err := strconv.Atoi(parts[1]); err == nil {
			c.Port = port
		}
	} else {
		c.Server = serverName
	}
	return c
}

func (c *SqlConnectionString) WithDatabase(database string) *SqlConnectionString {
	c.Database = database
	return c
}

func (c *SqlConnectionString) WithPort(port int) *SqlConnectionString {
	c.Port = port
	return c
}

func (c *SqlConnectionString) Valid() bool {
	if err := guard.EmptyOrNil(c.Database, "database"); err != nil {
		return false
	}

	if err := guard.EmptyOrNil(c.Password, "password"); err != nil {
		return false
	}

	if err := guard.EmptyOrNil(c.Server, "server"); err != nil {
		return false
	}

	if err := guard.EmptyOrNil(c.Username, "username"); err != nil {
		return false
	}

	return true
}

func (c *SqlConnectionString) Parse(connectionString string) error {
	userServerParts := strings.Split(connectionString, "@")
	if len(userServerParts) != 2 {
		return errors.New("wrong format, expecting user:password@tpc(server)/database")
	}
	userParts := strings.Split(userServerParts[0], ":")
	if len(userParts) != 2 {
		return errors.New("wrong format, expecting user:password@tpc(server)/database")
	}
	c.Username = strings.TrimSpace(userParts[0])
	c.Password = strings.TrimSpace(userParts[1])

	urlPart := ""
	databasePart := ""
	parametersPart := ""
	if strings.Contains(userServerParts[1], "?") {
		parametersPart = strings.Split(userServerParts[1], "?")[1]
		urlPart = strings.Split(userServerParts[1], "?")[0]
	}

	if urlPart == "" {
		urlPart = userServerParts[1]
	}
	if strings.Contains(urlPart, "/") {
		databasePart = strings.Split(urlPart, "/")[1]
		urlPart = strings.Split(urlPart, "/")[0]
	}
	if urlPart == "" {
		urlPart = userServerParts[1]
	}

	server := urlPart
	server = strings.ReplaceAll(server, "tcp(", "")
	server = strings.ReplaceAll(server, ")", "")
	serverParts := strings.Split(server, ":")
	if len(serverParts) == 1 {
		c.Server = serverParts[0]
	} else {
		c.Server = serverParts[0]
		if port, err := strconv.Atoi(serverParts[1]); err == nil {
			c.Port = port
		}
	}

	if databasePart != "" {
		c.Database = databasePart
	}
	if parametersPart != "" {
		parameters := strings.Split(parametersPart, "&")
		for _, parameter := range parameters {
			if parameter == "tls=true" {
				c.EnableTLS = true
			}
		}
	}

	return nil
}
