package account

import (
	"testing"

	"github.com/tsuru/config"
	. "gopkg.in/check.v1"

	"github.com/albertoleal/backstage/db"
	"github.com/albertoleal/backstage/errors"
)

type S struct{}

var _ = Suite(&S{})

//Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

func (s *S) SetUpSuite(c *C) {
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "backstage_db_test")
}

func (s *S) TearDownSuite(c *C) {
	conn, err := db.Conn()
	c.Assert(err, IsNil)
	defer conn.Close()
	config.Unset("database:url")
	config.Unset("database:name")
	// conn.Collection("services").Database.DropDatabase()
}

func (s *S) TestCreateServiceNewService(c *C) {
	user := User{Username: "alice"}
	service := Service{Name: "Backstage",
		Endpoint:  map[string]interface{}{"latest": "http://example.org/api"},
		Subdomain: "BACKSTAGE",
	}
	err := CreateService(&service, &user)
	defer DeleteService(&service)

	c.Check(service.Subdomain, Equals, "backstage")
	_, ok := err.(*errors.ValidationError)
	c.Check(ok, Equals, false)
}

func (s *S) TestCannotCreateServiceServiceWhenSubdomainAlreadyExists(c *C) {
	user := User{Username: "alice"}
	service := Service{Subdomain: "backstage",
		Endpoint: map[string]interface{}{"latest": "http://example.org/api"},
	}
	err := CreateService(&service, &user)
	defer DeleteService(&service)
	c.Check(err, IsNil)

	service2 := Service{Subdomain: "backstage",
		Endpoint: map[string]interface{}{"latest": "http://example.org/api"},
	}
	err = CreateService(&service2, &user)
	c.Check(err, NotNil)

	e, ok := err.(*errors.ValidationError)
	c.Assert(ok, Equals, true)
	message := "There is another service with this subdomain."
	c.Assert(e.Message, Equals, message)
}

func (s *S) TestCannotCreateServiceAServiceWithoutRequiredFields(c *C) {
	user := User{Username: "alice"}
	service := Service{Subdomain: "backstage"}
	err := CreateService(&service, &user)
	e := err.(*errors.ValidationError)
	message := "Endpoint cannot be empty."
	c.Assert(e.Message, Equals, message)

	service = Service{}
	err = CreateService(&service, &user)
	e = err.(*errors.ValidationError)
	message = "Subdomain cannot be empty."
	c.Assert(e.Message, Equals, message)
}

func (s *S) TestDeleteServiceANonExistingService(c *C) {
	service := Service{Subdomain: "backstage",
		Endpoint: map[string]interface{}{"latest": "http://example.org/api"},
	}
	err := DeleteService(&service)

	e, ok := err.(*errors.ValidationError)
	c.Assert(ok, Equals, true)
	message := "Document not found."
	c.Assert(e.Message, Equals, message)
}

func (s *S) TestDeleteServiceAnExistingService(c *C) {
	user := User{Username: "alice"}
	service := Service{Subdomain: "backstage",
		Endpoint: map[string]interface{}{"latest": "http://example.org/api"},
	}

	count, _ := CountService()
	c.Assert(count, Equals, 0)

	CreateService(&service, &user)
	count, _ = CountService()
	c.Assert(count, Equals, 1)

	DeleteService(&service)
	count, _ = CountService()
	c.Assert(count, Equals, 0)
}
