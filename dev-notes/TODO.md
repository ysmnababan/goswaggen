# TODO

This is the To-do lists for the `goswaggen` project.

- [X] Check function name
- [X] Learn the best method to test the `ast` package (unit or integration test)
- [X] Investigate how to recognize a handler (echo, gin, or net/http)
- [X] Recognize for all the handler in one file
- [X] Create method for associate the function and its comment block
- [X] How to parse the function param (other than`&{echo Context}`)
- [X] Learn how to use `go/packages`
- [ ] Experiment with `go/types` object.
- [ ] Find handler declaration position
- [ ] Get all function that use import
- [ ] Look up what param need to cover
- [ ] Ensure the package has correct import lib for the handler
- [ ] Learn how traverse to another file (from import)
- [ ] Learn how to fetch the router
- [ ] Learn how to fetch the success and failure response
- [ ] Learn how to add to recognize the payload, param or query param
- [ ] Learn how to update the comment without changing important field
- [ ] Add unit test for each function