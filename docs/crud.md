Common CRUD Design in Go
========================

[

Ben Johnson

20 Jan 2021

](https://web.archive.org/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)

*   [](https://web.archive.org/web/20250915154003/https://twitter.com/share?text=Common%20CRUD%20Design%20in%20Go&url=https://www.gobeyond.dev/crud/ "Share on Twitter")
*   [](https://web.archive.org/web/20250915154003/https://www.facebook.com/sharer/sharer.php?u=https://www.gobeyond.dev/crud/ "Share on Facebook")
*   [](https://web.archive.org/web/20250915154003/https://www.linkedin.com/shareArticle?mini=true&url=https://www.gobeyond.dev/crud/&title=Common%20CRUD%20Design%20in%20Go "Share on LinkedIn")
*   [](https://web.archive.org/web/20250915154003/mailto:/?subject=Common%20CRUD%20Design%20in%20Go&body=https://www.gobeyond.dev/crud/ "Share via Email")
*   [](# "Copy link")

Create, Read, Update, & Delete (CRUD) is the tech industry's bread-and-butter. You're familiar with it if you've spent any time doing application development.

Many programming languages lean on frameworks to provide an opinionated structure for CRUD applications, but the Go community is notoriously anti-framework. As such, we need to have our own CRUD design.

After years of developing Go applications, I've found a common design that has worked well across different projects. We'll be looking at the [WTF Dial](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf?ref=gobeyond.dev) project as an example. You can read more about that project in [this introductory blog post](https://web.archive.org/web/20250915154003/https://www.gobeyond.dev/wtf-dial/).

The interface
-------------

In WTF Dial, we define our services with an interface in the root package which represents our business domain. This allows us to create different implementations that share a common contract. In [dial.go](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev), we define the `wtf.DialService` interface:

    type DialService interface {
    	FindDialByID(ctx context.Context, id int) (*Dial, error)
    	FindDials(ctx context.Context, filter DialFilter) ([]*Dial, int, error)
    	CreateDial(ctx context.Context, dial *Dial) error
    	UpdateDial(ctx context.Context, id int, upd DialUpdate) (*Dial, error)
    	DeleteDial(ctx context.Context, id int) error
    }

[dial.go#L81-L122](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev#L81-L122)

This structure is what I use for nearly all entities across my applications. It provides a simple structure, but it's flexible enough to work in most cases.

### Transactional boundaries

I view my service definitions as a black box. As such, I rarely expose internal details like transactions to the rest of my application. While it might be tempting to let the caller of your service compose individual transactional calls, it's rarely necessary and typically complicates your application.

### Enforcing security through context

In WTF Dial, users are authenticated when a request comes in and the authenticated user object is added to the `context.Context` via the `[NewContextWithUser()](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/context.go?ref=gobeyond.dev#L24-L27)` function. This means that the current user is available to any function in our service via the `ctx` argument.

Authorization enforcement is built into the service implementation for a few reasons. First, it ensures that restrictions are enforced at the lowest level possible instead of delegating to a higher level abstraction. It's less likely that we forget a security check when we can embed it directly in our SQL query. Second, it can be more efficient to push these restrictions to the database layer as it limits the data queried and returned.

Here is an example of a security check within the `sqlite.findDials()` function where we limit the query to only dials that the user is a member of:

    userID := wtf.UserIDFromContext(ctx)
    where = append(where, `id IN (SELECT dial_id FROM dial_memberships dm WHERE dm.user_id = ?)`)
    args = append(args, userID)

[sqlite/dial.go#L306-L310](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/45d236828b5140980286d0e01f996810e5f4b99d/sqlite/dial.go?ref=gobeyond.dev#L306-L310)

Looking up individual objects
-----------------------------

Finding an object by primary key is one of the most common tasks you'll encounter. Here we define a function for fetching a `wtf.Dial` by its `id`:

    FindDialByID(ctx context.Context, id int) (*Dial, error)

This function definition looks deceptively simple but there are important semantics to determine. What happens if the dial isn't found? What related data do we return with the dial?

### Don't return double nil

A common pitfall I see is that developers will return a `nil` dial and `nil` error if the ID cannot be found. In this context, however, a user is expecting a specific dial so not finding it would be an error condition.

In practice, callers to the function will perform a simple `err != nil` check but it's easy to forget to check for a `nil` dial as well. This will cause your program to panic.

    // Try to fetch the dial by ID but it doesn't exist.
    dial, err := FindDialByID(ctx, 100)
    if err != nil {
    	return err
    }
    
    // Oops! Panic here because dial is nil.
    fmt.Printf("WTF Level: %d", dial.Value)

Always return either the object or an error; both should not be nil

### Choosing what data you return

When returning our Dial object, the caller typically needs related information as well. Who owns the dial? Who are other members of the dial? Our data is a graph that can branch out infinitely so we need to enforce a boundary.

We could allow the caller to define the graph using GraphQL or even just a set of flags but that adds complexity to our application. It's easier to return a generally useful set of related data instead. We will incur extra database calls or increased network bandwidth but that's usually a good trade-off at first and we can optimize use cases as needed.

I typically return related data which has a parent relationship to the main object. In the case of `Dial`, it has a `User` parent object that I'll attach. These relationships are almost always required by the caller because they give context to the object.

![](https://web.archive.org/web/20250915154003im_/https://www.gobeyond.dev/content/images/2021/01/Common-CRUD--1-.svg)

Adding child relationship can easily explode the object graph

I will include child relationships if I know there will be a limited number of child objects and that they will almost always be fetched when viewing the parent object. In the case of `Dial`, we could include the list of members of the dial because that is typically useful and we'll never have more than a handful of members. Another good example would be returning a set of order items with an e-commerce order.

Searching for multiple objects
------------------------------

Our next function provides a way to search for dials by a variety of filtering options. Fetching a list of dials sounds similar to fetching a single dial but there are some important differences.

    FindDials(ctx context.Context, filter DialFilter) ([]*Dial, int, error)

### Returning double nil is ok

Unlike our `FindDialByID()`, it's ok to return no dials and to return a `nil` error. The caller likely doesn't know if there should be any matching dials—that's why they're searching—so not matching any dials is not an error condition.

We also don't need to worry about panicking like we did when searching for a single dial because we are returning a slice. Most operations on a slice (`len()` or `for in`) will work fine on a `nil` slice value.

    // Search for a list of all dials.
    dials, _, err := FindDials(ctx, DialFilter{})
    if err != nil {
    	return err
    }
    
    // No panic this time. A nil slice of dials is ok.
    fmt.Printf("You have %d dials.", len(dials))

Returning a nil list and a nil error will not cause a panic

### Filtering results

In this function, we pass in a `filter` object instead of multiple filtering arguments. This allows us to add additional filters without breaking API compatibility in the future.

    // DialFilter represents a filter used by FindDials().
    type DialFilter struct {
    	// Filtering fields.
    	ID         *int    `json:"id"`
    	InviteCode *string `json:"inviteCode"`
    
    	// Restrict to subset of range.
    	Offset int `json:"offset"`
    	Limit  int `json:"limit"`
    }

[dial.go#L124-L133](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/dial.go?ref=gobeyond.dev#L124-L133)

We're using pointers in our filter struct so that we can optionally add filters. Each field we set will further restrict the results.

![](https://web.archive.org/web/20250915154003im_/https://www.gobeyond.dev/content/images/2021/01/Common-CRUD---Filter--2-.svg)

Adding additional filter fields is analogous to AND-ing them together

### Slicing results & returning totals

The `Offset` & `Limit` fields in our `DialFilter` object above can be used to return a subset of the results and are analogous to the SQL `OFFSET` & `LIMIT` clauses.

However, it's still useful to know the total number of matching dials even if we limit the number of dials returned. For example, pagination requires the total count. To do this, we return an `int` in addition to our `[]*Dial` slice.

Some databases allow us to compute this in one SQL query using `COUNT(*) OVER()`. For example, if we are searching for dials with a user ID of 100 and we limit our search to 20 records, we can still get the total count like this:

    SELECT id, name, COUNT(*) OVER()
    FROM dials
    WHERE user_id = 100
    ORDER BY id
    LIMIT 20

Return up to 20 dials as well as the total count

We can iterate over our result set and extract the dial data as well as the total count like this:

    var dials []*Dial
    var n int
    for rows.Next() {
    	var dial Dial
    	if rows.Scan(&dial.ID, &dial.UserID, &n); err != nil {
    		return err
    	}
    	dials = append(dials, &dial)
    }

The total count (`n`) is scanned on each row but the value is the same every time

### Sorting results

As for sorting, you you don't want to allow users to sort by any column in your database. Most columns will not be indexed so the query will be slow. Instead, I recommend mapping a fixed set of values to your columns. For example, `"name_asc"` can map to an `ORDER BY name ASC` clause.

You can find an example of this in the WTF Dial when searching for memberships:

    var sortBy string
    switch filter.SortBy {
    case "updated_at_desc":
    	sortBy = "dm.updated_at DESC"
    default:
    	sortBy = `dm.ID ASC`
    }
    

[sqlite/dial\_membership.go#L164-L172](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/blob/e23f5f00e0f48f54bd751cc264ea85c094f7d466/sqlite/dial_membership.go?ref=gobeyond.dev#L164-L172)

In this snippet, we are checking if the `filter.SortBy` field is set to a predefined sort order ( `"updated_at_desc"`). If so, we translate that to a SQL snippet. Otherwise, we use a default sorting case.

Creating dials
--------------

To create a new user in our application, we have the following function:

    CreateDial(ctx context.Context, dial *Dial) error

Here we pass in the `Dial` object we want to create. We need to communicate the new dial ID back to the caller so we'll update the primary key (`dial.ID`) and any other fields generated by the service implementation (such as a creation date).

You can also return a separate `Dial` object from the function if you don't want to update the original. However, I've found this approach more cumbersome in practice.

### Transactionally building the object graph

Because we're restricting the transaction boundary to our function call, we should allow creation of nested objects as appropriate. For example, we could accept a list of `DialMembership` objects attached to the `Dial` that would be created in the same transaction.

    svc.CreateDial(ctx, &wtf.Dial{
    	UserID:      100,
    	Name:        "My Dial",
    	Memberships: []*wtf.DialMembership{
    		{
    			User:  &wtf.User{Email:"susy@que.com"},
    			Value: 50,
    		},
    		{
    			User:  &wtf.User{Email:"john@doe.com"},
    			Value: 50,
    		},
    	},
    })

Creating a dial and with a multiple members in one call

Updating existing dials
-----------------------

For updating existing users, we have the following function:

    UpdateDial(ctx context.Context, id int, upd DialUpdate) (*Dial, error)

This function updates a dial with a given ID with the field values set in `upd`. The newly updated dial is returned. Our update type, `DialUpdate`, lets us restrict our updates to a subset of fields:

    // DialUpdate represents a set of fields to update on a dial.
    type DialUpdate struct {
    	Name *string `json:"name"`
    }
    

Note that the pointer in the `Name` field indicates that it's optional. If it's unset, then it is not updated. Our `DialUpdate` type is simple, but we could imagine adding a `UserID` field if we wished to allow users to reassign the dial to someone else. This lets us avoid adding a new `ReassignDial()` to our service.

### Returning the dial on error

Unlike many Go functions, the `UpdateDial()` always returns a dial object even when an error has occurred. This is useful because the user typically wants to see the state they attempted to update the dial to if there was a validation error. It's especially important for web-based applications where the each HTTP request is stateless.

### Bulk update

The `id` field is intentionally separated from the `DialUpdate` type so we could allow bulk updates as well. For example, we could build a function called `UpdateDials()`:

    UpdateDials(ctx context.Context, ids []int, upd DialUpdate) ([]*Dial, error)

By changing the function to accept a list of IDs, we can apply it to all of them. In turn, we now return a list of updated dials.

Deleting dials
--------------

Honestly, there's not much to say about deletions. We have a simple function to delete by primary key:

    DeleteDial(ctx context.Context, id int) error

We can expand this into a bulk delete by providing a slice of IDs:

    DeleteDials(ctx context.Context, id []int) error

Be sure to enforce authorization restrictions to ensure a user can't delete another user's dial.

Conclusion
----------

Optimizing your CRUD application development is crucial as it makes up a majority of most application code. We've taken a look at a basic framework for structuring your Go CRUD functions that balances the trade-off of flexibility and simplicity. You'll need to tweak this framework as every application has its own unique requirements but hopefully it starts you off with a solid foundation.

If you have questions, comments, or suggestions, please visit the [WTF Dial GitHub Discussion board](https://web.archive.org/web/20250915154003/https://github.com/benbjohnson/wtf/discussions?ref=gobeyond.dev). There's already been some great posts and it's provided a much better place to discuss as compared to traditional comment sections.

[Application Design](/web/20250915154003/https://www.gobeyond.dev/tag/application-design/)

[](/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)

### [Ben Johnson](/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)[

](https://web.archive.org/web/20250915154003/https://x.com/benbjohnson)

Freelance Go developer, author of BoltDB

* * *

### You might also like

[Real-world SQL in Go: Part I](/web/20250915154003/https://www.gobeyond.dev/real-world-sql-part-one/)
-----------------------------------------------------------------------------------------------------

Regardless of whether you hate SQL or merely tolerate it, you're going to use it in a project at some point. Relational database structures don't always map well to application data structures but SQL's ubiquity means that it's the standard tool developers

[

Ben Johnson

26 Jan 2021

](/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)

[Application Design](/web/20250915154003/https://www.gobeyond.dev/tag/application-design/)

[The Go Object Lifecycle](/web/20250915154003/https://www.gobeyond.dev/the-go-object-lifecycle/)
------------------------------------------------------------------------------------------------

Despite such a simple language, Go developers have found a surprising number of ways to create and use objects. In this post we’ll look at a 3-step approach to object management—instantiation, initialization, & initiation. We’ll also contrast this with other methodologies for creating and using objects and

[

Ben Johnson

20 Jun 2018

](/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)

[Application Design](/web/20250915154003/https://www.gobeyond.dev/tag/application-design/)

*   [](https://web.archive.org/web/20250915154003/https://x.com/gobeyonddev)
*   [](https://web.archive.org/web/20250915154003/https://www.facebook.com/Go-Beyond-103788501641633)
*   [](/web/20250915154003/https://www.gobeyond.dev/rss "RSS")

### Featured Posts

[

### Standard Package Layout

10 Aug 2016

](/web/20250915154003/https://www.gobeyond.dev/standard-package-layout/)

### [Authors →](https://web.archive.org/web/20250915154003/https://www.gobeyond.dev/authors)

[

### Ben Johnson

Freelance Go developer, author of BoltDB



](/web/20250915154003/https://www.gobeyond.dev/author/benbjohnson/)
