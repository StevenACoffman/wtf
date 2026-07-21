Real-world SQL in Go: Part I
============================

[

Ben Johnson

26 Jan 2021

](https://web.archive.org/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)

*   [](https://web.archive.org/web/20250809025020/https://twitter.com/share?text=Real-world%20SQL%20in%20Go%3A%20Part%20I&url=https://www.gobeyond.dev/real-world-sql-part-one/ "Share on Twitter")
*   [](https://web.archive.org/web/20250809025020/https://www.facebook.com/sharer/sharer.php?u=https://www.gobeyond.dev/real-world-sql-part-one/ "Share on Facebook")
*   [](https://web.archive.org/web/20250809025020/https://www.linkedin.com/shareArticle?mini=true&url=https://www.gobeyond.dev/real-world-sql-part-one/&title=Real-world%20SQL%20in%20Go%3A%20Part%20I "Share on LinkedIn")
*   [](https://web.archive.org/web/20250809025020/mailto:/?subject=Real-world%20SQL%20in%20Go%3A%20Part%20I&body=https://www.gobeyond.dev/real-world-sql-part-one/ "Share via Email")
*   [](# "Copy link")

Regardless of whether you hate SQL or merely tolerate it, you're going to use it in a project at some point. Relational database structures don't always map well to application data structures but SQL's ubiquity means that it's the standard tool developers reach for when adding data persistence.

While Go has object-relational mapping libraries available such as [GORM](https://web.archive.org/web/20250809025020/https://gorm.io/?ref=gobeyond.dev), we'll eschew that abstraction and integrate directly with the `database/sql` standard library package. ORM tools add a layer of complexity that can help with simple data access patterns but make advanced queries & debugging more difficult.

In this post, we'll be looking at how to physically structure data access code, where to place transactional boundaries, and a few useful SQL tricks. I'll be referencing code from the [WTF Dial project](https://web.archive.org/web/20250809025020/https://wtfdial.com/?ref=gobeyond.dev). You can read about it with this [introductory blog post](https://web.archive.org/web/20250809025020/https://www.gobeyond.dev/wtf-dial/) to get some context.

Implementing our service interface
----------------------------------

In the previous [blog post about CRUD design](https://web.archive.org/web/20250809025020/https://www.gobeyond.dev/crud/), I described an interface for managing `Dial` entities. A dial is a way to measure the level of frustration of a group and is owned by a user but can also have multiple inputs from other users which are called memberships in the dial.

    type DialService interface {
    	FindDialByID(ctx context.Context, id int) (*Dial, error)
    	FindDials(ctx context.Context, filter DialFilter) ([]*Dial, int, error)
    	CreateDial(ctx context.Context, dial *Dial) error
    	UpdateDial(ctx context.Context, id int, upd DialUpdate) (*Dial, error)
    	DeleteDial(ctx context.Context, id int) error
    }

[dial.go#L81-L122](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev#L81-L122)

We'll be implementing this `wtf.DialService` for SQLite so it will be in the `sqlite` subpackage and it will be called `sqlite.DialService`. The concepts we'll talk about here will work for any SQL database though.

### Abstracting the database

It may seem odd to name our package something generic like `sqlite` but our package will be wrapping the underlying `database/sql` and `[mattn/go-sqlite3](https://web.archive.org/web/20250809025020/https://github.com/mattn/go-sqlite3?ref=gobeyond.dev)` packages and will not expose the internals of those packages. The caller of our package will only interact with SQL & SQLite through our package. You could name it `wtfsqlite` or `wtfsql` but I personally find that ugly.

Since we are not exposing the underlying SQL packages, we'll have our own `DB` type in our `sqlite` package:

    // DB represents a database connection to our application.
    type DB struct {
    	db     *sql.DB
    
    	// Datasource name.
    	DSN string
    }
    
    // Open opens the database connection.
    func (db *DB) Open() (err error)
    
    // Close closes the database connection.
    func (db *DB) Close() error

[sqlite/sqlite.go#L43-L58](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/sqlite.go?ref=gobeyond.dev#L43-L58)

This abstraction lets us handle the setup & teardown of the database in our `[Open()](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/sqlite.go?ref=gobeyond.dev#L72-L113)` & `Close()` methods without requiring the caller package to worry about the details.

### Separating interface implementation from SQL helpers

Within my `sqlite` package, I've found it useful to make a distinction between the service interface implementation and the functions that actually execute SQL. The service implementation exists for a two reasons:

*   Provide an implementation of the `wtf.DialService` interface.
*   Provide a transactional boundary.

As such, the service methods are typically small and in this format:

    func (s *DialService) CreateDial(ctx context.Context, dial *wtf.Dial) error {
    	tx, err := s.db.BeginTx(ctx, nil)
    	if err != nil {
    		return err
    	}
    	defer tx.Rollback()
    
    	// Call helper functions...
        
    	return tx.Commit()
    }
    

[sqlite/dial.go#L75-L99](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L75-L99)

In this example, we start a transaction, execute some lower level helper functions, and then commit. If any error occurs in our helper functions then the deferred rollback will abort the changes. If our function reaches the end and successfully calls `tx.Commit()` then the deferred rollback will be a no-op.

A helper function typically takes this form:

    func createDial(ctx context.Context, tx *Tx, dial *wtf.Dial) error

[sqlite/dial.go#L360-L418](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L360-L418)

The helper functions are where the functionality lives and they are not attached to a particular service. They can be reused by different services or even called by other helper functions. It's an easy way to abstract your higher level code from the low-level SQL calls.

Implementing entity search
--------------------------

Our `wtf.DialService` interface defines a function for searching for dials called `FindDials()`:

    package wtf
    
    type DialService interface {
    	FindDials(ctx context.Context, filter DialFilter) ([]*Dial, int, error)
    	...
    }
    

[dial.go#L81-L122](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev#L81-L122)

### Service implementation

The implementation of  service method in our `sqlite` package look like this:

    func (s *DialService) FindDials(ctx context.Context, filter wtf.DialFilter) ([]*wtf.Dial, int, error) {
    	tx, err := s.db.BeginTx(ctx, nil)
    	if err != nil {
    		return nil, 0, err
    	}
    	defer tx.Rollback()
    
    	// Fetch list of matching dial objects.
    	dials, n, err := findDials(ctx, tx, filter)
    	if err != nil {
    		return dials, n, err
    	}
    
    	// Iterate over dials and attach associated owner user.
    	// This should be batched up if using a remote database server.
    	for _, dial := range dials {
    		if err := attachDialAssociations(ctx, tx, dial); err != nil {
    			return dials, n, err
    		}
    	}
    	return dials, n, nil
    }

[sqlite/dial.go#L47-L73](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L47-L73)

From a high level, it performs three steps:

1.  Establish the transactional boundary.
2.  Find a list of matching dials.
3.  Find associated data for each dial (such as the dial owner).

Working with associated data (aka joins) is a complicated topic and I'll be covering that in the next post.

### Helper methods

We have a helper method for `Dial` search because we want to reuse it for other services. For example, when looking up a user we may want to also return a list of dials for that user.

Our helper method looks like this from a high level:

    func findDials(ctx context.Context, tx *Tx, filter wtf.DialFilter) ([]*wtf.Dial, int, error) {
    	// Build WHERE clause...
    
    	// Execute query.
    	rows, err := tx.QueryContext(ctx, `SELECT ...`)
    	if err != nil {
    		return nil, n, err
    	}
    	defer rows.Close()
    
    	// Iterate over rows and deserialize into Dial objects.
    	dials := make([]*wtf.Dial, 0)
    	for rows.Next() {
    		var dial wtf.Dial
    		if rows.Scan(&dial.ID, ...); err != nil {
    			return nil, 0, err
    		}
    		dials = append(dials, &dial)
    	}
    	return dials, n, rows.Err()
    }

[sqlite/dial.go#L292-L358](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L292-L358)

Essentially, we're trying to convert SQL fields into our application domain object, `wtf.Dial`. There are a couple interesting details though.

#### SQL builders vs hand-written SQL

Even if you're not using an ORM, there are options such as [Squirrel](https://web.archive.org/web/20250809025020/https://github.com/Masterminds/squirrel?ref=gobeyond.dev) for building SQL queries. I haven't personally used Squirrel but it looks like it could be a good option. However, for most of my queries I find that hand-rolling SQL works pretty well.

The `WHERE` clause is typically the most dynamic part of the query so we'll focus on `WHERE` clause generation. Our application domain specifies a `wtf.DialFilter` that is a set of filter fields that are `AND`\-ed together. We perform this by building a slice of conditions:

    // List of WHERE clause predicates. Start with an always true predicate
    // so we have at least one when joining later.
    where := []string{"1 = 1"}
    
    // Bind parameters used by the WHERE clause. 
    var args []interface{}{}
    
    // Add predicate for ID if specified on the filter.
    if v := filter.ID; v != nil {
    	where = append(where, "id = ?")
    	args = append(args, *v)
    }

[sqlite/dial.go#L295-L311](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L295-L311)

Then when we build our query we can use `strings.Join()`:

    whereClause := strings.Join(where, " AND ")
    q := `SELECT ... FROM dials WHERE ` + whereClause

Joining predicates to create a WHERE clause

If we don't have any filters then our clause becomes `WHERE 1 = 1` which matches all rows. If we add our ID filter then it becomes `WHERE 1 = 1 AND id = ?`. The SQL query optimizer is smart enough to ignore an always true condition like `1 = 1`.

### Obtaining total row count

Our service interface requires that `FindDials()` returns the total number of matching dials to use for pagination. For example, if a `Limit` of `20` is set but there are a hundred matching dials then the caller wants to know that so they can show that there are 5 pages of results.

The naive approach is to execute your query and then re-execute it but with only a `COUNT(1)` and no `OFFSET` or `LIMIT` clause. However, there's some SQL trickery that lets you combine these into one query.

Many databases such as SQLite or Postgres provide windowing functions. We can specify a window as one of our fields that will utilize the `WHERE` clause but ignore the `OFFSET`/`LIMIT`:

    SELECT id, name, COUNT(*) OVER()
    FROM dials
    WHERE user_id = ?
    OFFSET 40 LIMIT 20

Calculating the total matching count in your original query

The `COUNT(*) OVER ()` field specifies that it should return the total count of the query over an empty window. It will return the same value for every row of our query so we can simply `Scan()` it into a count variable on every iteration.

Implementing entity lookup
--------------------------

We could implement entity look up with `FindDials()` but it's such a common task that I find it useful to create a separate function called `FindDialByID()`:

    package wtf
    
    type DialService interface {
    	FindDialByID(ctx context.Context, id int) (*Dial, error)
    	...
    }
    

[dial.go#L81-L122](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev#L81-L122)

### Service implementation

Our implementation in our service looks similar to `FindDials()` except that we are now looking up a single dial:

    func (s *DialService) FindDialByID(ctx context.Context, id int) (*wtf.Dial, error) {
    	tx, err := s.db.BeginTx(ctx, nil)
    	if err != nil {
    		return nil, err
    	}
    	defer tx.Rollback()
    
    	// Fetch dial object and attach owner user.
    	dial, err := findDialByID(ctx, tx, id)
    	if err != nil {
    		return nil, err
    	} else if err := attachDialAssociations(ctx, tx, dial); err != nil {
    		return nil, err
    	}
    
    	return dial, nil
    }

[sqlite/dial.go#L26-L45](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L26-L45)

### Helper function

Our underlying `findDialByID()` actually wraps the `findDials()` function but it changes the semantics:

    // findDialByID is a helper function to retrieve a dial by ID.
    // Returns ENOTFOUND if dial doesn't exist.
    func findDialByID(ctx context.Context, tx *Tx, id int) (*wtf.Dial, error) {
    	dials, _, err := findDials(ctx, tx, wtf.DialFilter{ID: &id})
    	if err != nil {
    		return nil, err
    	} else if len(dials) == 0 {
    		return nil, &wtf.Error{Code: wtf.ENOTFOUND, Message: "Dial not found."}
    	}
    	return dials[0], nil
    }

[sqlite/dial.go#L265-L275](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/dial.go?ref=gobeyond.dev#L265-L275)

Whereas the `findDials()` will not return an error if no data is found, the `findDialByID()` function expects that the given dial exists. Here we translate an empty result set into a _"not found"_ error.

You can find details about using an application-specific `Error` type in the post, [Failure is Your Domain](https://web.archive.org/web/20250809025020/https://www.gobeyond.dev/failure-is-your-domain/). I'll be writing an updated version of that soon that includes the Go 1.13 error wrapping changes but most of the concepts still hold true.

### Optimizing primary key lookups

It might be tempting to create a separate optimized function for looking up dials by ID instead of reusing code from `findDials()`. It would allow you to skip a slice allocation and optimize the query field list to include related object data such as the owner information from the `users` table.

However, duplicating your query code makes it more difficult to maintain. It's best to wait until you find hot paths in your code that require optimization and only optimize them.

Conclusion
----------

In this post we've looked at implementing the read side of our domain's service interfaces and we've separated out our high-level service methods from lower-level SQL helper functions. We also took a look at optimizing our pagination queries by integrating with SQL windowing functions. In the next post, we'll look at implementing the write side of our domain service.

It's easy to have SQL permeate throughout your application if you're not careful. By isolating our database code to a single package, we ensure that the rest of our application code can grow cleanly & independently.

Have comments or suggestions about this post or future content you'd like to see? I've set up a [GitHub Discussion board](https://web.archive.org/web/20250809025020/https://github.com/benbjohnson/wtf/discussions?ref=gobeyond.dev) instead of using a comments system so that folks can have deeper conversations.

[Application Design](/web/20250809025020/https://www.gobeyond.dev/tag/application-design/)

[](/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)

### [Ben Johnson](/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)[

](https://web.archive.org/web/20250809025020/https://x.com/benbjohnson)

Freelance Go developer, author of BoltDB

* * *

### You might also like

[Common CRUD Design in Go](/web/20250809025020/https://www.gobeyond.dev/crud/)
------------------------------------------------------------------------------

Create, Read, Update, & Delete (CRUD) is the tech industry's bread-and-butter. You're familiar with it if you've spent any time doing application development. Many programming languages lean on frameworks to provide an opinionated structure for CRUD applications, but the Go community is notoriously anti-framework.

[

Ben Johnson

20 Jan 2021

](/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)

[Application Design](/web/20250809025020/https://www.gobeyond.dev/tag/application-design/)

[The Go Object Lifecycle](/web/20250809025020/https://www.gobeyond.dev/the-go-object-lifecycle/)
------------------------------------------------------------------------------------------------

Despite such a simple language, Go developers have found a surprising number of ways to create and use objects. In this post we’ll look at a 3-step approach to object management—instantiation, initialization, & initiation. We’ll also contrast this with other methodologies for creating and using objects and

[

Ben Johnson

20 Jun 2018

](/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)

[Application Design](/web/20250809025020/https://www.gobeyond.dev/tag/application-design/)

*   [](https://web.archive.org/web/20250809025020/https://x.com/gobeyonddev)
*   [](https://web.archive.org/web/20250809025020/https://www.facebook.com/Go-Beyond-103788501641633)
*   [](/web/20250809025020/https://www.gobeyond.dev/rss "RSS")

### Featured Posts

[

### Standard Package Layout

10 Aug 2016

](/web/20250809025020/https://www.gobeyond.dev/standard-package-layout/)

### [Authors →](https://web.archive.org/web/20250809025020/https://www.gobeyond.dev/authors)

[

### Ben Johnson

Freelance Go developer, author of BoltDB



](/web/20250809025020/https://www.gobeyond.dev/author/benbjohnson/)
