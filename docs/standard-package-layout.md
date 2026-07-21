Standard Package Layout
=======================

[

Ben Johnson

10 Aug 2016

](https://web.archive.org/web/web/20250915154151/https://www.gobeyond.dev/author/benbjohnson/)

*   [](https://web.archive.org/web/20250915154151/https://twitter.com/share?text=Standard%20Package%20Layout&url=https://www.gobeyond.dev/standard-package-layout/ "Share on Twitter")
*   [](https://web.archive.org/web/20250915154151/https://www.facebook.com/sharer/sharer.php?u=https://www.gobeyond.dev/standard-package-layout/ "Share on Facebook")
*   [](https://web.archive.org/web/20250915154151/https://www.linkedin.com/shareArticle?mini=true&url=https://www.gobeyond.dev/standard-package-layout/&title=Standard%20Package%20Layout "Share on LinkedIn")
*   [](https://web.archive.org/web/20250915154151/mailto:/?subject=Standard%20Package%20Layout&body=https://www.gobeyond.dev/standard-package-layout/ "Share via Email")
*   [](# "Copy link")

Vendoring. Generics. These are seen as big issues in the Go community but there’s another issue that’s rarely mentioned — application package layout.

Every Go application I’ve ever worked on appears to have a different answer to the question, __how should I organize my code?__ Some applications push everything into one package while others group by type or module. Without a good strategy applied across your team, you’ll find code scattered across various packages of your application. We need a better standard for Go application design.

I suggest a better approach. By following a few simple rules we can decouple our code, make it easier to test, and bring a consistent structure to our project. Before we dive into it, though, let’s look at some of the most common ways people structure projects today.

Common flawed approaches
========================

There seem to be a handful of common approaches to Go application organization that each have their own flaws.

Approach #1: Monolithic package
-------------------------------

Throwing all your code in a single package can actually work very well for small applications. It removes any chance of circular dependencies because, within your application, there are no dependencies.

I’ve seen this work for applications up to 10K [SLOC](https://web.archive.org/web/20250915154151/https://en.wikipedia.org/wiki/Source_lines_of_code?ref=gobeyond.dev). Beyond that size, it gets extremely difficult to navigate the code and isolate your code.

Approach #2: Rails-style layout
-------------------------------

Another approach is to group your code by it’s functional type. For example, all your [handlers](https://web.archive.org/web/20250915154151/https://golang.org/pkg/net/http/?ref=gobeyond.dev#Handler) go in one package, your controllers go in another, and your models go in yet another. I see this a lot from former [Rails](https://web.archive.org/web/20250915154151/http://rubyonrails.org/?ref=gobeyond.dev) developers (myself included).

There are two issues with this approach though. First, your names are atrocious. You end up with type names like __controller.UserController__ where you’re duplicating your package name in your type‘s name. I tend to be a stickler about naming. I believe your names are your best documentation when you’re down in the weeds coding. Names are also used as a proxy for quality — it’s the first thing someone notices when reading code.

The bigger issue, however, is circular dependencies. Your different functional types may need to reference each other. This __only__ works if you have one-way dependencies but many times your application is not that simple.

Approach #3: Group by module
----------------------------

This approach is similar to the Rails-style layout except that we are grouping our code by module instead of by function. For example, you may have a __users__ package and an __accounts__ package.

We find the same issues in this approach. Again, we end up with terrible names like __users.User.__ We also have the same issue of circular dependencies if our __accounts.Controller__ needs to interact with our __users.Controller__ and vis-a-versa.

A better approach
=================

The package strategy that I use for my projects involves 4 simple tenets:

1.  Root package is for domain types
2.  Group subpackages by dependency
3.  Use a shared __mock__ subpackage
4.  __Main__ package ties together dependencies

These rules help isolate our packages and define a clear domain language across the entire application. Let’s look at how each one of these rules works in practice.

#1. Root package is for domain types
====================================

Your application has a logical, high-level language that describes how data and processes interact. This is your domain. If you have an e-commerce application your domain involves things like customers, accounts, charging credit cards, and handling inventory. If you’re Facebook then your domain is users, likes, & relationships. It’s the stuff that doesn’t depend on your underlying technology.

I place my domain types in my root package. This package only contains simple data types like a __User__ struct for holding user data or a __UserService__ interface for fetching or saving user data.

It may look something like:

    package myapp
    
    type User struct {
    	ID      int
    	Name    string
    	Address Address
    }
    
    type UserService interface {
    	User(id int) (*User, error)
    	Users() ([]*User, error)
    	CreateUser(u *User) error
    	DeleteUser(id int) error
    }

This makes your root package extremely simple. You may also include types that perform actions but only if they solely depend on other domain types. For example, you could have a type that polls your __UserService__ periodically. However, it should not call out to external services or save to a database. That is an implementation detail.

__The root package should not depend on any other package in your application!__

#2. Group subpackages by dependency
===================================

If your root package is not allowed to have external dependencies then we must push those dependencies to subpackages. In this approach to package layout, subpackages exist as an adapter between your domain and your implementation.

For example, your __UserService__ might be backed by PostgreSQL. You can introduce a __postgres__ subpackage in your application that provides a __postgres.UserService__ implementation:

    package postgres
    
    import (
    	"database/sql"
    
    	"github.com/benbjohnson/myapp"
    	_ "github.com/lib/pq"
    )
    
    // UserService represents a PostgreSQL implementation of myapp.UserService.
    type UserService struct {
    	DB *sql.DB
    }
    
    // User returns a user for a given id.
    func (s *UserService) User(id int) (*myapp.User, error) {
    	var u myapp.User
    	row := db.QueryRow(`SELECT id, name FROM users WHERE id = $1`, id)
    	if row.Scan(&u.ID, &u.Name); err != nil {
    		return nil, err
    	}
    	return &u, nil
    }
    
    // implement remaining myapp.UserService interface...

This isolates our PostgreSQL dependency which simplifies testing and provides an easy way to migrate to another database in the future. It can be used as a pluggable architecture if you decide to support other database implementations such as [BoltDB](https://web.archive.org/web/20250915154151/https://github.com/boltdb/bolt?ref=gobeyond.dev).

It also gives you a way to layer implementations. Perhaps you want to hold an in-memory, [LRU cache](https://web.archive.org/web/20250915154151/https://en.wikipedia.org/wiki/Cache_algorithms?ref=gobeyond.dev) in front of PostgreSQL. You can add a __UserCache__ that implements __UserService__ which can wrap your PostgreSQL implementation:

    package myapp
    
    // UserCache wraps a UserService to provide an in-memory cache.
    type UserCache struct {
            cache   map[int]*User
            service UserService
    }
    
    // NewUserCache returns a new read-through cache for service.
    func NewUserCache(service UserService) *UserCache {
            return &UserCache{
                    cache: make(map[int]*User),
                    service: service,
            }
    }
    
    // User returns a user for a given id.
    // Returns the cached instance if available.
    func (c *UserCache) User(id int) (*User, error) {
    	// Check the local cache first.
            if u := c.cache[id]]; u != nil {
                    return u, nil
            }
    
    	// Otherwise fetch from the underlying service.
            u, err := c.service.User(id)
            if err != nil {
            	return nil, err
            } else if u != nil {
            	c.cache[id] = u
            }
            return u, err
    }

We see this approach in the standard library too. The __io.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/io/?ref=gobeyond.dev#Reader) is a domain type for reading bytes and its implementations are grouped by dependency — __tar.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/archive/tar/?ref=gobeyond.dev#Reader), __gzip.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/compress/gzip/?ref=gobeyond.dev#Reader), __multipart.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/mime/multipart/?ref=gobeyond.dev#Reader). These can be layered as well. It’s common to see an __os.__[__File__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/os/?ref=gobeyond.dev#File) wrapped by a __bufio.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/bufio/?ref=gobeyond.dev#Reader) which is wrapped by a __gzip.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/compress/gzip/?ref=gobeyond.dev#Reader) which is wrapped by a __tar.__[__Reader__](https://web.archive.org/web/20250915154151/https://golang.org/pkg/archive/tar/?ref=gobeyond.dev#Reader).

Dependencies between dependencies
---------------------------------

Your dependencies don’t live in isolation. You may store __User__ data in PostgreSQL but your financial transaction data exists in a third party service like [Stripe](https://web.archive.org/web/20250915154151/https://stripe.com/?ref=gobeyond.dev). In this case we wrap our Stripe dependency with a logical domain type — let’s call it __TransactionService__.

By adding our __TransactionService__ to our __UserService__ we decouple our two dependencies:

    type UserService struct {
    	DB                 *sql.DB
    	TransactionService myapp.TransactionService
    }

Now our dependencies communicate solely through our common domain language. This means that we could swap out PostgreSQL for MySQL or switch Stripe for another payment processor without affecting other dependencies.

Don’t limit this to third party dependencies
--------------------------------------------

This may sound odd but I also isolate my standard library dependencies with this same method. For instance, the __net/http__ package is just another dependency. We can isolate it as well by including an __http__ subpackage in our application.

It might seem odd to have a package with the same name as the dependency it wraps, however, this is intentional. There are no package name conflicts in your application unless you allow __net/http__ to be used in other parts of your application. The benefit to duplicating the name is that it requires you to isolate all HTTP code to your __http__ package.

    package http
    
    import (
            "net/http"
            
            "github.com/benbjohnson/myapp"
    )
    
    type Handler struct {
            UserService myapp.UserService
    }
    
    func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
            // handle request
    }

Now your __http.Handler__ acts as an adapter between your domain and the HTTP protocol.

#3. Use a shared mock subpackage
================================

Because our dependencies are isolated from other dependencies by our domain interfaces, we can use these connection points to inject mock implementations.

There are several mocking libraries such as [GoMock](https://web.archive.org/web/20250915154151/https://github.com/golang/mock?ref=gobeyond.dev) that will generate mocks for you but I personally prefer to just write them myself. I find many of the mocking tools to be overly complicated.

The mocks I use are very simple. For example, a mock for the __UserService__ looks like:

    package mock
    
    import "github.com/benbjohnson/myapp"
    
    // UserService represents a mock implementation of myapp.UserService.
    type UserService struct {
            UserFn      func(id int) (*myapp.User, error)
            UserInvoked bool
    
            UsersFn     func() ([]*myapp.User, error)
            UsersInvoked bool
    
            // additional function implementations...
    }
    
    // User invokes the mock implementation and marks the function as invoked.
    func (s *UserService) User(id int) (*myapp.User, error) {
            s.UserInvoked = true
            return s.UserFn(id)
    }
    
    // additional functions: Users(), CreateUser(), DeleteUser()

This mock lets me inject functions into anything that uses the __myapp.UserService__ interface to validate arguments, return expected data, or inject failures.

Let’s say we want to test our __http.Handler__ that we built above:

    package http_test
    
    import (
    	"testing"
    	"net/http"
    	"net/http/httptest"
    
    	"github.com/benbjohnson/myapp/mock"
    )
    
    func TestHandler(t *testing.T) {
    	// Inject our mock into our handler.
    	var us mock.UserService
    	var h Handler
    	h.UserService = &us
    
    	// Mock our User() call.
    	us.UserFn = func(id int) (*myapp.User, error) {
    		if id != 100 {
    			t.Fatalf("unexpected id: %d", id)
    		}
    		return &myapp.User{ID: 100, Name: "susy"}, nil
    	}
    
    	// Invoke the handler.
    	w := httptest.NewRecorder()
    	r, _ := http.NewRequest("GET", "/users/100", nil)
    	h.ServeHTTP(w, r)
    	
    	// Validate mock.
    	if !us.UserInvoked {
    		t.Fatal("expected User() to be invoked")
    	}
    }

Our mock lets us completely isolate our unit test to only the handling of the HTTP protocol.

#4. Main package ties together dependencies
===========================================

With all these dependency packages floating around in isolation, you may wonder how they all come together. That’s the job of the __main__ package.

Main package layout
-------------------

An application may produce multiple binaries so we’ll use the Go convention of placing our __main__ package as a subdirectory of the __cmd package.__ For example, our project may have a __myapp__ server binary but also a __myappctl__ client binary for managing the server from the terminal. We’ll layout our main packages like this:

    myapp/
        cmd/
            myapp/
                main.go
            myappctl/
                main.go

Injecting dependencies at compile time
--------------------------------------

The term “dependency injection” has gotten a bad rap. It conjures up thoughts of verbose [Spring](https://web.archive.org/web/20250915154151/https://projects.spring.io/spring-framework/?ref=gobeyond.dev) XML files. However, all the term really means is that we’re going to pass dependencies to our objects instead of requiring that the object build or find the dependency itself.

The __main__ package is what gets to choose which dependencies to inject into which objects. Because the __main__ package simply wires up the pieces, it tends to be fairly small and trivial code:

    package main
    
    import (
    	"log"
    	"os"
    	
    	"github.com/benbjohnson/myapp"
    	"github.com/benbjohnson/myapp/postgres"
    	"github.com/benbjohnson/myapp/http"
    )
    
    func main() {
    	// Connect to database.
    	db, err := postgres.Open(os.Getenv("DB"))
    	if err != nil {
    		log.Fatal(err)
    	}
    	defer db.Close()
    
    	// Create services.
    	us := &postgres.UserService{DB: db}
    
    	// Attach to HTTP handler.
    	var h http.Handler
    	h.UserService = us
    	
    	// start http server...
    }

It’s also important to note that your __main__ package is also an adapter. It connects the terminal to your domain.

Conclusion
==========

Application design is a hard problem. There are so many design decisions to make and without a set of solid principles to guide you the problem is made even worse. We’ve looked at several current approaches to Go application design and we’ve seen many of their flaws.

I believe approaching design from the standpoint of dependencies makes code organization simpler and easier to reason about. First we design our domain language. Then we isolate our dependencies. Next we introduce mocks to isolate our tests. Finally, we tie everything together within our __main__ package.

Consider these principles in the next application you design. If you have any questions or want to discuss design, contact me at [@benbjohnson](https://web.archive.org/web/20250915154151/https://twitter.com/benbjohnson?ref=gobeyond.dev) on Twitter or find me as __benbjohnson__ on the [Gopher slack](https://web.archive.org/web/20250915154151/https://gophersinvite.herokuapp.com/?ref=gobeyond.dev).

[Application Design](/web/20250915154151/https://www.gobeyond.dev/tag/application-design/)

[](/web/20250915154151/https://www.gobeyond.dev/author/benbjohnson/)

### [Ben Johnson](/web/20250915154151/https://www.gobeyond.dev/author/benbjohnson/)[

](https://web.archive.org/web/20250915154151/https://x.com/benbjohnson)

Freelance Go developer, author of BoltDB

* * *
