# FAQ & Troubleshooting

### Why isn't anything being generated?
- Make sure you're pointing to the correct folder (e.g., `goswaggen ./controllers`).
- Ensure your handler functions are **exported** and attached to a receiver (e.g., `func (c *UserController)`).
- Only `.go` files are processed — files with build tags may be skipped.

### Can I customize the generated annotations?
Not yet. Template customization is planned in a future release (see roadmap in GitHub issues).

### Does it support frameworks other than Echo?
Currently supports:
- Echo (`echo.Context`)
- Gin (`*gin.Context`) — beta

We’ll consider Chi or Fiber if demand grows.

### I already have comments. Will they be overwritten?
Only if the comment block matches known patterns. Otherwise, the new block is inserted above the function without deleting anything. Use `--dry-run` to preview.

### Can I undo the changes?
Use version control! Always `git commit` before running the tool, or use `--dry-run` to inspect.
