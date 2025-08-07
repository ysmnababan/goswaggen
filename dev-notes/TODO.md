# TODO

This is the To-do lists for the `goswaggen` project.

DEVELOPMENT STAGE:Analyze all the relevant information before generating comment
- [X] Check function name
- [X] Learn the best method to test the `ast` package (unit or integration test)
- [X] Investigate how to recognize a handler (echo, gin, or net/http)
- [X] Recognize for all the handler in one file
- [X] Create method for associate the function and its comment block
- [X] How to parse the function param (other than`&{echo Context}`)
- [X] Learn how to use `go/packages`
- [X] Experiment with `go/types` object.
- [X] Find handler declaration position for simple registration
- [X] Get all function that use import
- [X] find the handler for group handler registration
- [X] refactor the program with `(n * node)`
- [X] make sure the no duplicate handler node captured
- [X] Find the 'not direct' handler registration
- [X] Learn how to fetch the router
- [X] Learn how traverse to another file (from import)
- [X] Reinspect the handler using previously fetched `types.Func`
- [X] Ensure the package has correct import lib for the handler
- [X] Relation between the `c.Bind` with the http method
- [X] Rule or convention for the http method and the payload, param, or query param
- [X] Add cache for storing global or const variable
- [X] Learn how to add to recognize the payload, param or query param
- [X] Traverse the `Bind()` param (as struct) and retrieve the field with associated tag
- [X] Check how to find the http response from the `return` statement
- [X] Learn how to fetch the success and failure response
- [X] Check the `@Produce` type for each `echo` response from documentation
- [X] Create unit test for the parser
- [X] Refactor 

DEVELOPMENT STAGE: Generating Comment
- [ ] Look up what param need to cover
- [ ] Learn how to update the comment without changing important field



DEVELOPMENT STAGE: Rewriting the Comment
- [ ] Learn how to insert comment to existing file


GENERAL:
- [ ] Add unit test for each function
- [ ] Adding command line capability