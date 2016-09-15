# structfix

updates omitted or updated struct definitions in nested struct elements
a workaround for https://github.com/golang/go/issues/6064

you can add it to your editors onsave actions

example:

```golang
type foo struct{
	bar struct {
		baz int
	}
}

var f = foo{bar:{baz:42}}
```

```bash
$ structfix -w .
```


```golang
type foo struct{
	bar struct {
		baz int
	}
}

var f = foo{bar:struct {
		baz int
	}{baz:42}}
```




