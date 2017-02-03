# Gestalt

Gestalt is an integration test framework built for automating
CLI-based workflows.

## Example

 1. Create a server
 1. Launch an app on server.
 1. Create a user on the application.

## Components

### gestalt.Component

```go
func Login() gestalt.Component {
  return gestalt.NewComponent("login",func(e gestalt.Evaluator) result.Result {
    
  }
}
```

### component.BG

### component.Retry

### component.Ensure

### component.Group

### component.Suite

```go
func Login() gestalt.Component {
  return
}
```
