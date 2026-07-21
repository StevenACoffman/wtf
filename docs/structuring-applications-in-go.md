Structuring Applications in Go
==============================

[

Ben Johnson

06 Jul 2014

](https://web.archive.org/web/20250915153044/https://www.gobeyond.dev/author/benbjohnson/)

*   [](https://web.archive.org/web/20250915153044/https://twitter.com/share?text=Structuring%20Applications%20in%20Go&url=https://www.gobeyond.dev/structuring-applications/ "Share on Twitter")
*   [](https://web.archive.org/web/20250915153044/https://www.facebook.com/sharer/sharer.php?u=https://www.gobeyond.dev/structuring-applications/ "Share on Facebook")
*   [](https://web.archive.org/web/20250915153044/https://www.linkedin.com/shareArticle?mini=true&url=https://www.gobeyond.dev/structuring-applications/&title=Structuring%20Applications%20in%20Go "Share on LinkedIn")
*   [](https://web.archive.org/web/20250915153044/mailto:/?subject=Structuring%20Applications%20in%20Go&body=https://www.gobeyond.dev/structuring-applications/ "Share via Email")
*   [](# "Copy link")

For me, the hardest part of learning Go was in structuring my application. Prior to Go, I was working on a Rails application and Rails makes you structure your application in a certain way. “Convention over configuration” is their motto. But Go doesn’t prescribe any particular project layout or application structure and Go’s conventions are mostly stylistic.

I’m going to show you four patterns that I’ve found to be tremendously helpful in architecting Go applications. These are not official Gopher rules and I’m sure others may have differing opinions. I’d love to hear them! Please comment as you go through if you have suggestions.

1\. Don’t use global variables
==============================

The Go net/http examples I read always show a function registered with [http.HandleFunc](https://web.archive.org/web/20250915153044/http://golang.org/pkg/net/http/?ref=gobeyond.dev#HandleFunc) like this:

    package main
    
    import (
    	“fmt”
    	“net/http”
    )
    
    func main() {
    	http.HandleFunc(“/hello”, hello)
    	http.ListenAndServe(“:8080", nil)
    }
    
    func hello(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, “hi!”)
    }

This example gives an easy way to get into using net/http but it teaches a bad habit. By using a function handler, the only way to access application state is to use a global variable. Because of this, you may decide to add a global database connection or a global configuration variable but these globals are a nightmare to use when writing unit tests.

A better way is to make specific types for handlers so they can include the required variables:

    type HelloHandler struct {
    	db *sql.DB
    }
    
    func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    	var name string
    
    	// Execute the query.
    	row := h.db.QueryRow(“SELECT myname FROM mytable”)
    	if err := row.Scan(&name); err != nil {
    		http.Error(w, err.Error(), 500)
    		return
    	}
    
    	// Write it back to the client.
    	fmt.Fprintf(w, “hi %s!\n”, name)
    }

Now we can initialize our database and register our handler without the use of global variables:

    func main() {
        // Open our database connection.
        db, err := sql.Open(“postgres”, “…”)
        if err != nil {
            log.Fatal(err)
        }
    
        // Register our handler.
        http.Handle(“/hello”, &HelloHandler{db: db})
        http.ListenAndServe(“:8080", nil)
    }

This approach also has the benefit that unit testing our handler is self contained and doesn’t even require an HTTP server:

    func TestHelloHandler_ServeHTTP(t *testing.T) {
        // Open our connection and setup our handler.
        db, _ := sql.Open("postgres", "...")
        defer db.Close()
        h := HelloHandler{db: db}
    
        // Execute our handler with a simple buffer.
        rec := httptest.NewRecorder()
        rec.Body = bytes.NewBuffer()
        h.ServeHTTP(rec, nil)
    
        if rec.Body.String() != "hi bob!\n" {
            t.Errorf("unexpected response: %s", rec.Body.String())
        }
    }

****__UPDATE:__**** Tomás Senart and Peter Bourgon [mentioned on Twitter](https://web.archive.org/web/20250915153044/https://twitter.com/tsenart/status/485920561391239168?ref=gobeyond.dev) that you can simplify this further by [wrapping your handlers with a closure](https://web.archive.org/web/20250915153044/https://gist.github.com/tsenart/5fc18c659814c078378d?ref=gobeyond.dev). This allows you to easily compose your handlers.

2\. Separate your binary from your application
==============================================

I used to place my __main.go__ file in the root of my project so that when someone runs “go get” then my application would be automagically installed. However, combining the __main.go__ file and my application logic in the same package has two consequences:

1.  It makes my application unusable as a library.
2.  I can only have one application binary.

The best way I’ve found to fix this is to simply use a __“cmd”__ directory in my project where each of its subdirectories is an application binary. I originally found this approach used in Brad Fitzpatrick’s [Camlistore](https://web.archive.org/web/20250915153044/http://camlistore.org/?ref=gobeyond.dev) project where he uses several application binaries:

    camlistore/
        cmd/
            camget/
                main.go
            cammount/
                main.go
            camput/
                main.go
            camtool/
                main.go

Here we have 4 separate application binaries that can be built when Camlistore is installed: `camget`, `cammount`, `camput`, & `camtool`.

Library driven development
--------------------------

Moving the __main.go__ file out of your root allows you to build your application from the perspective of a library. Your application binary is simply a client of your application’s library. I find this helps me make a cleaner abstraction of what code is for my core logic (the library) and what code is for running my application (the application binary).

The application binary is really just the interface for how a user interacts with your logic. Sometimes you might want users to interact in multiple ways so you create multiple binaries. For example, if you had an “adder” package that that let users add numbers together, you may want to release a command line version as well as a web version. You can easily do this by organizing your project like this:

    adder/
        adder.go
        cmd/
            adder/
                main.go
            adder-server/
                main.go

Users can install your “adder” application binaries with “go get” using an ellipsis:$ go get github.com/benbjohnson/adder/...

And voila, your user has __“adder”__ and __“adder-server”__ installed!

3\. Wrap types for application-specific context
===============================================

One trick I’ve found especially helpful is realizing that some generic types should be wrapped to provide application-level context. A great example of this is wrapping the DB and Tx (transaction) types. These types can be found in the database/sql package or other database libraries such as [Bolt](https://web.archive.org/web/20250915153044/https://github.com/boltdb/bolt?ref=gobeyond.dev).

We start by wrapping these types like this:

    package myapp
    
    import (
    	"database/sql"
    )
    
    type DB struct {
    	*sql.DB
    }
    
    type Tx struct {
    	*sql.Tx
    }

We then wrap the initialization function for our database and transaction:

    // Open returns a DB reference for a data source.
    func Open(dataSourceName string) (*DB, error) {
        db, err := sql.Open("postgres", dataSourceName)
        if err != nil {
            return nil, err
        }
        return &DB{db}, nil
    }
    
    // Begin starts an returns a new transaction.
    func (db *DB) Begin() (*Tx, error) {
        tx, err := db.DB.Begin()
        if err != nil {
            return nil, err
        }
        return &Tx{tx}, nil
    }

And now we can add application specific functions to our transactions. For example, if our application has users that need to be validated before being created, a __Tx.CreateUser()__ would be a good function to add:

    // CreateUser creates a new user.
    // Returns an error if user is invalid or the tx fails.
    func (tx *Tx) CreateUser(u *User) error {
    	// Validate the input.
    	if u == nil {
    		return errors.New("user required")
    	} else if u.Name == "" {
    		return errors.New("name required")
    	}
    
    	// Perform the actual insert and return any errors.
    	return tx.Exec(`INSERT INTO users (...) VALUES`, ...)
    }

This function can get more complicated if, for example, a user needs to be validated against another system before being created or other tables need to be updated. To your application’s caller, though, it’s all isolated in one function.

Transactional composition
-------------------------

Another benefit to adding these functions to your __Tx__ is that it allows you to compose multiple actions in a single transaction. Need to add one user? Just call __Tx.CreateUser()__ once:

    tx, _ := db.Begin()
    tx.CreateUser(&User{Name:"susy"})
    tx.Commit()
    

Need to add a bunch of users? You can use the same function. No need for a __Tx.CreateUsers()__ function:

    tx, _ := db.Begin()
    for _, u := range users {
    	tx.CreateUser(u)
    }
    tx.Commit()

Abstracting your underlying data store also makes it trivial to swap out a new database or to use multiple databases. They’re all hidden from your calling code by your application’s __DB__ & __Tx__ types.

4\. Don’t go crazy with subpackages
===================================

Most languages let you organize your package structure however you’d like. I’ve worked in Java codebases where every couple of classes get stuffed into another package and these packages would all include each other. It was a mess!

Go only has one requirement for packages: you can’t have cyclic dependencies. This cyclic dependency rule felt strange to me at first. I originally organized my project so each file had one type and once there were a bunch of files in a package then I’d create a new subpackage. However, these subpackages became difficult to manage since I couldn’t have package “A” include package “B” which included package “C” which included package “A”. That would be a cyclic dependency. I realized that I had no good reason for separating out packages except for having “too many files”.

Recently I’ve found myself going the other direction — only using a single root package. Usually my project’s types are all very related so it fits better from a usability and API standpoint. These types can also take advantage of calling unexported between them which keeps the API small and clear.

I found a few things helped me move toward larger packages:

1.  Group related types and code together in each file. If your types and functions are well organized then I find that files tend to be between 200 and 500 SLOC. This might sound like a lot but I find it easy to navigate. 1000 SLOC is usually my upper limit for a single file.
2.  Organize the most important type at the top of the file and add types in decreasing importance towards the bottom of the file.
3.  Once your application starts getting above 10,000 SLOC you should seriously evaluate whether it can be broken into smaller projects.

[Bolt](https://web.archive.org/web/20250915153044/https://github.com/boltdb/bolt?ref=gobeyond.dev) is a good example of this. Each file is a grouping of types related to a single Bolt construct:

    bucket.go
    cursor.go
    db.go
    freelist.go
    node.go
    page.go
    tx.go

Conclusion
==========

Code organization is one of the hardest parts about writing software and it rarely gets the focus it deserves. Use global variables sparingly, move your application binary code to its own package, wrap some types for application-specific context, and limit your subpackages. These are just a few tricks that can help make Go code easier and more maintainable.

If you’re writing Go projects the same way you write Ruby, Java, or Node.js projects then you’re probably going to be fighting with the language.

[Application Design](/web/20250915153044/https://www.gobeyond.dev/tag/application-design/)

[](/web/20250915153044/https://www.gobeyond.dev/author/benbjohnson/)

### [Ben Johnson](/web/20250915153044/https://www.gobeyond.dev/author/benbjohnson/)[

](https://web.archive.org/web/20250915153044/https://x.com/benbjohnson)

Freelance Go developer, author of BoltDB

* * *
