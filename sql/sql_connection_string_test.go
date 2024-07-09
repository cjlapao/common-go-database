package sql

import (
	"reflect"
	"testing"
)

func TestSqlConnectionString_Parse(t *testing.T) {
	type args struct {
		connectionString string
	}
	type expect struct {
		server   string
		port     int
		database string
		user     string
		password string
		tls      bool
	}
	tests := []struct {
		name    string
		c       *SqlConnectionString
		args    args
		expect  expect
		wantErr bool
	}{
		{
			name: "without database",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1)",
			},
			expect: expect{
				database: "",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     0,
				tls:      false,
			},
			wantErr: false,
		},
		{
			name: "without database and with port",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1:20)",
			},
			expect: expect{
				database: "",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     20,
				tls:      false,
			},
			wantErr: false,
		},
		{
			name: "with database and port",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1:20)/test",
			},
			expect: expect{
				database: "test",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     20,
				tls:      false,
			},
			wantErr: false,
		},
		{
			name: "with database and tls",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1)/test?tls=true",
			},
			expect: expect{
				database: "test",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     0,
				tls:      true,
			},
			wantErr: false,
		},
		{
			name: "with database and tls false",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1)/test?tls=false",
			},
			expect: expect{
				database: "test",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     0,
				tls:      false,
			},
			wantErr: false,
		},
		{
			name: "with database and wrong parameter",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1)/test?something=false",
			},
			expect: expect{
				database: "test",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     0,
				tls:      false,
			},
			wantErr: false,
		},
		{
			name: "with database and no port",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "admin:test@tcp(127.0.0.1)/test",
			},
			expect: expect{
				database: "test",
				user:     "admin",
				password: "test",
				server:   "127.0.0.1",
				port:     0,
			},
			wantErr: false,
		},
		{
			name: "missing user database",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "test@tcp(db.babyready.io)",
			},
			expect:  expect{},
			wantErr: true,
		},
		{
			name: "missing user database",
			c:    &SqlConnectionString{},
			args: args{
				connectionString: "test_tcp(db.babyready.io)",
			},
			expect:  expect{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Parse(tt.args.connectionString); (err != nil) != tt.wantErr {
				t.Errorf("SqlConnectionString.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.expect.database != tt.c.Database {
					t.Errorf("expected database to be %v found %v", tt.expect.database, tt.c.Database)
				}
				if tt.expect.server != tt.c.Server {
					t.Errorf("expected server to be %v found %v", tt.expect.server, tt.c.Server)
				}
				if tt.expect.port != tt.c.Port {
					t.Errorf("expected port to be %v found %v", tt.expect.port, tt.c.Port)
				}
				if tt.expect.user != tt.c.Username {
					t.Errorf("expected user to be %v found %v", tt.expect.user, tt.c.Username)
				}
				if tt.expect.password != tt.c.Password {
					t.Errorf("expected password to be %v found %v", tt.expect.password, tt.c.Password)
				}
				if tt.expect.tls != tt.c.EnableTLS {
					t.Errorf("expected tls to be %v found %v", tt.expect.tls, tt.c.EnableTLS)
				}
			}
		})
	}
}

func TestSqlConnectionString_Valid(t *testing.T) {
	tests := []struct {
		name string
		c    *SqlConnectionString
		want bool
	}{
		{
			name: "invalid database",
			c: &SqlConnectionString{
				Username: "test",
				Password: "test",
				Port:     0,
				Server:   "test",
			},
			want: false,
		},
		{
			name: "invalid user",
			c: &SqlConnectionString{
				Database: "test",
				Password: "test",
				Port:     0,
				Server:   "test",
			},
			want: false,
		},
		{
			name: "invalid password",
			c: &SqlConnectionString{
				Database: "test",
				Username: "test",
				Port:     0,
				Server:   "test",
			},
			want: false,
		},
		{
			name: "invalid server",
			c: &SqlConnectionString{
				Database: "test",
				Username: "test",
				Password: "test",
				Port:     0,
			},
			want: false,
		},
		{
			name: "valid",
			c: &SqlConnectionString{
				Database: "test",
				Username: "test",
				Password: "test",
				Port:     0,
				Server:   "test",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Valid(); got != tt.want {
				t.Errorf("SqlConnectionString.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_WithUser(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name string
		c    *SqlConnectionString
		args args
		want *SqlConnectionString
	}{
		{
			name: "UserSet",
			c:    &SqlConnectionString{},
			args: args{
				username: "test",
			},
			want: &SqlConnectionString{
				Username: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.WithUser(tt.args.username); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SqlConnectionString.WithUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_WithPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		c    *SqlConnectionString
		args args
		want *SqlConnectionString
	}{
		{
			name: "Password Set",
			c:    &SqlConnectionString{},
			args: args{
				password: "test",
			},
			want: &SqlConnectionString{
				Password: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.WithPassword(tt.args.password); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SqlConnectionString.WithPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_WithServer(t *testing.T) {
	type args struct {
		server string
	}
	tests := []struct {
		name string
		c    *SqlConnectionString
		args args
		want *SqlConnectionString
	}{
		{
			name: "Server without port",
			c:    &SqlConnectionString{},
			args: args{
				server: "example.com",
			},
			want: &SqlConnectionString{
				Server: "example.com",
				Port:   0,
			},
		},
		{
			name: "Server with port",
			c:    &SqlConnectionString{},
			args: args{
				server: "example.com:3306",
			},
			want: &SqlConnectionString{
				Server: "example.com",
				Port:   3306,
			},
		},
		{
			name: "Server with invalid port",
			c:    &SqlConnectionString{},
			args: args{
				server: "example.com:abc",
			},
			want: &SqlConnectionString{
				Server: "example.com",
				Port:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.WithServer(tt.args.server); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SqlConnectionString.WithServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_WithDatabase(t *testing.T) {
	type args struct {
		database string
	}
	tests := []struct {
		name string
		c    *SqlConnectionString
		args args
		want *SqlConnectionString
	}{
		{
			name: "Database Set",
			c:    &SqlConnectionString{},
			args: args{
				database: "test",
			},
			want: &SqlConnectionString{
				Database: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.WithDatabase(tt.args.database); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SqlConnectionString.WithDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_WithPort(t *testing.T) {
	type args struct {
		port int
	}
	tests := []struct {
		name string
		c    *SqlConnectionString
		args args
		want *SqlConnectionString
	}{
		{
			name: "Port Set",
			c:    &SqlConnectionString{},
			args: args{
				port: 20,
			},
			want: &SqlConnectionString{
				Port: 20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.WithPort(tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SqlConnectionString.WithPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSqlConnectionString_ConnectionString(t *testing.T) {
	tests := []struct {
		name string
		c    *SqlConnectionString
		want string
	}{
		{
			name: "get connection string",
			c: &SqlConnectionString{
				Username: "test",
				Password: "test",
				Server:   "example.com",
				Database: "test",
				Port:     0,
			},
			want: "test:test@tcp(example.com)/test?parseTime=true",
		},
		{
			name: "get connection string with port",
			c: &SqlConnectionString{
				Username: "test",
				Password: "test",
				Server:   "example.com",
				Database: "test",
				Port:     3306,
			},
			want: "test:test@tcp(example.com:3306)/test?parseTime=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.ConnectionString(); got != tt.want {
				t.Errorf("SqlConnectionString.ConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}
