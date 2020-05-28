# godocage

A Go wrapper for the [Docage](http://api.docage.com/) API.

## Configuration

Import the package like so:

```go
import "github.com/georges760/godocage"
```

Then initiate an API struct with your credentials:

```go
//explicitly
api := gobcy.API{}
api.Email = "your-api-email-here"
api.Key = "your-api-key-here"

//using a struct literal
api := gobcy.API{"your-api-email-here","your-api-key-here"}

//query away
var contactid string
var contact Contact
contact.Email = "moi@free.fr"
contact.FirstName = "John"
contact.LastName = "Doe"
contact.Country = "France"
contactid, err = api.CreateContact(contact)
if err != nil {
	log.Fatal("Error creating contact: ", err)
}
log.Println("contactid:", contactid)
```

## Usage

You can check out [Docage's documentation here](https://documenter.getpostman.com/view/10595102/SzRuYXd1?version=latest). The `godocage_test.go` file also shows most of the API calls in action.

## Testing

The aforementioned `godocage_test.go` file contains a number of tests to ensure the wrapper is functioning properly. If you run it yourself, you'll have to insert a valid API token.

