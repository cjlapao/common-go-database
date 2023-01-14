package sqlite

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cjlapao/common-go/guard"
	"github.com/cjlapao/common-go/helper"
)

type SqlConnectionString struct {
	FilePath string
	FileName string
}

func (c *SqlConnectionString) ConnectionString() string {
	if c.FilePath == "" {
		c.FilePath = "./"
	}

	if !strings.HasSuffix(c.FileName, ".db") {
		c.FileName = fmt.Sprintf("%s.db", c.FileName)
	}

	return helper.JoinPath(c.FilePath, c.FileName)
}

func (c *SqlConnectionString) WithDatabase(database string) *SqlConnectionString {
	if !strings.HasSuffix(database, ".db") {
		database = fmt.Sprintf("%s.db", database)
	}

	c.FileName = database
	return c
}

func (c *SqlConnectionString) WithPath(path string) *SqlConnectionString {
	if path == "" {
		path = "./"
	}
	c.FilePath = path
	return c
}

func (c *SqlConnectionString) Valid() bool {
	if err := guard.EmptyOrNil(c.FileName, "database"); err != nil {
		return false
	}

	return true
}

func (c *SqlConnectionString) Parse(path string) error {
	if path == "" {
		return fmt.Errorf("Path cannot be empty and need to be a valid filepath")
	}

	dir := filepath.Dir(path)
	file := filepath.Base(path)

	c.FileName = file
	c.FilePath = dir

	return nil
}
