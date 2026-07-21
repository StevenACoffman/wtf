The Purpose of Tests
--------------------

People argue over testing style, whether to use TDD or BDD, or whether tests are even useful at all. Before I get into how I structure my tests in Go, I should explain how I see tests.

Tests should be 2 things:

1.  Self-contained
2.  Easily reproducible

That’s it. They should be self-contained so that changing one part of your test suite doesn’t drastically affect another part. They should be easily reproducible so that someone doesn’t have to go through multiple steps to get their test suite running the same as mine.

With that explanation, here are some of my own rules for testing in Go.

1\. Don’t use frameworks
------------------------

It seems like everyone has their own testing framework for Go. Some do setup/teardown, some do BDD-style function chaining, & some do a web interface. Don’t use them. Go has a perfectly good testing framework built in. Frameworks are also one more barrier to entry for other developers contributing to your code.

I do find that Go’s testing framework makes test assertions verbose so I add three simple helper functions ([code here](https://github.com/benbjohnson/testing)):
```
func assert(tb testing.TB, condition **bool**, msg **string**)   
func ok(tb testing.TB, err **error**)   
func equals(tb testing.TB, exp, act **interface**{}) 
```
These functions consolidate the typical call/check-error/check-value format of tests into an easier to digest format. So instead of:
```
func TestSomething(t *testing.T) {  
    value, err := DoSomething()  
    if err != nil {  
        t.Fatalf("DoSomething() failed: %s", err)  
    }  
    if value != 100 {  
        t.Fatalf("expected 100, got: %d", value)  
    }  
}
```
We can simply write:
```
func TestSomething(t \*testing.T) {  
    value, err := DoSomething()  
    ok(t, err)  
    equals(t, 100, value)  
}
```
That’s a lot clearer to me. The functions are also in my package so I don’t have to “go get” any dependencies.

2\. Use the “underscore test” package
-------------------------------------

Go doesn’t let you include multiple packages in one folder except for when you use a separate test package. For example, we can have these two files in one folder:

**user.go:**
```
package myapptype User struct {  
    id int  
    Name string  
}func (u *User) Save() error {  
    if u.id == 0 {  
        return u.create()  
    }  
    return u.update()  
}func (u *User) create() error { ... }  
func (u *User) update() error { ... }
```
**user\_test.go:**
```
package myapp_test
import (  
    "testing"    . 
    "github.com/benbjohnson/myapp"  
)
func TestUser\_Save(t *testing.T) {  
    u := &User{Name: "Susy Queue"}  
    ok(t, u.Save())  
}
```
In this example, my User type can perform a save but whether it creates or updates the user record in the database is internal to my package and my test shouldn’t have to worry about it.

Using a separate “myapp\_test” package means that I can’t access the unexported fields and functions in my package. I find this lets me test as a user of the package so I can see if my exported API is usable and complete. When I test in the same package I tend to muck around with the internals of my package which makes my tests brittle.

_UPDATE: Dave Cheney brought up a_ [_good point on Twitter_](https://twitter.com/davecheney/status/489395485136809984) _about the use of the “dot” import. Since this test is testing the exported API, it doesn’t make sense to pretend to be inside of the package. I like that approach and I’ll use it in the future._

3\. Use test-specific types
---------------------------

I talked previously about how [I wrap some generic types like sql.DB & sql.Tx](/@benbjohnson/structuring-applications-in-go-3b04be4ff091) with my own application-specific types to provide functions specific to my application. In my tests, I do something similar so that I can have functions specific to my tests.

[](/plans?source=promotion_paragraph---post_body_banner_rabbit_hole_blocks--46ddee7a25c---------------------------------------)

Most of my applications center around a DB type so let’s use that as an example. Let’s say I have an application specific DB type that wraps a [Bolt](https://github.com/boltdb/bolt) database:
```
type DB struct {  
    \*bolt.DB  
}func Open(path string, mode os.FileMode) (\*DB, error) {  
    db, err := bolt.Open(path, mode)  
    if err != nil {  
        return nil, err  
    }  
    return &DB{db}, nil  
}
```
In my test, I’ll add a TestDB type that automatically opens my database with a temporary file and then provides a close function that I can use for clean up:
```
type TestDB struct {  
    \*DB  
}// NewTestDB returns a TestDB using a temporary path.  
func NewTestDB() \*TestDB {  
    // Retrieve a temporary path.  
    f, err := ioutil.TempFile("", "")  
    if err != nil {  
        panic("temp file: %s", err)  
    }  
    path := f.Name()  
    f.Close()  
    os.Remove(path)    // Open the database.  
    db, err := Open(path, 0600)  
    if err != nil {  
        panic("open: %s", err)  
    }    // Return wrapped type.  
    return &TestDB{db}  
}// Close and delete Bolt database.  
func (db \*TestDB) Close() {  
    defer os.Remove(db.Path())  
    db.DB.Close()  
}
```
Now in our test it’s easy to do setup and teardown in 2 lines:
```
func TestDB\_DoSomething(t \*testing.T) {  
    db := NewTestDB()  
    defer db.Close()  
    ...  
}
```
We can also add other dependencies to our test type so that everything sets up and tears down together.

4\. Use inline interfaces & simple mocks
----------------------------------------

I’ve gone back and forth on best practices for interfaces. Initially I used them frequently but they made my code more complex. Then I stopped using them altogether and it made it more difficult to mock external dependencies.

The biggest turning point for me was realizing that my caller should create the interface instead of the callee providing an interface. This makes sense because the caller can declare exactly what it needs.

Let’s look at an example. Say we are using a third party client for everyone’s favorite social network: [Yo](http://www.justyo.co/). Our client looks like this:
```go
package yotype Client struct {}// Send sends a "yo" to someone.  
func (c *Client) Send(recipient string) error// Yos retrieves a list of my yo's.  
func (c *Client) Yos() ([]*Yo, error)
```
If my application only cares about sending yo’s then it can declare that with an inline interface:
```
package myapptype MyApplication struct {  
    YoClient interface {  
        Send(string) error  
    }  
}func (a *MyApplication) Yo(recipient string) error {  
    return a.YoClient.Send(recipient)  
}
```
Now in our main.go we can initialize the application and set the client:
```
package mainfunc main() {  
    c := yo.NewClient()  
    a := myapp.MyApplication{}  
    a.YoClient = c  
    ...  
}
```
And in our test we can use a mock implementation:
```
// package myapp_test TestYoClient provides mockable implementation of yo.Client.  
type TestYoClient struct {  
    SendFunc func(string) error  
}func (c *TestYoClient) Send(recipient string) error {  
    return c.SendFunc(recipient)  
}func TestMyApplication_SendYo(t *testing.T) {  
    c := &TestYoClient{}  
    a := &MyApplication{YoClient: c}    // Mock our send function to capture the argument.  
    var recipient string  
    c.SendFunc = func(s string) error {  
        recipient = s  
        return nil  
    }    // Send the yo and verify the recipient.  
    err := a.Yo("susy")  
    ok(t, err)  
    equals(t, "susy", recipient)  
}
```
This is a contrived example that simply passes the recipient through to the client. However, this approach is helpful for testing more complex workflows and for testing errors returned from external dependencies.

You can also use this approach for mocking the file system by having an "os" interface that has functions such as Create() or Open().

Conclusion
----------

With testing built into the Go toolchain, it’s easy to get started and use. Most of the hurdles I've experienced have been stylistic but I find that there are usually simple, idiomatic approaches to Go testing.

Use the built-in constructs for testing and make use of types to clarify test workflow. Test your package’s API instead of its internals by using a separate test package, and finally, use inline interfaces to decouple your application from external dependencies. These are just a few ways that can help make your Go tests more clear and maintainable.
