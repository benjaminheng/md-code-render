# Example

## Normal mode

![render-48fb4eec0b5d357512d90010cdef8bff.svg](render-48fb4eec0b5d357512d90010cdef8bff.svg)

```dot render
digraph G {
    A -> B -> C
}
```

![render-48fb4eec0b5d357512d90010cdef8bff.svg](render-48fb4eec0b5d357512d90010cdef8bff.svg)

```dot render{"mode": "normal"}
digraph G {
    A -> B -> C
}
```

## Code collapsed mode

![render-48fb4eec0b5d357512d90010cdef8bff.svg](render-48fb4eec0b5d357512d90010cdef8bff.svg)

<details><summary>Source</summary>

```dot render{"mode": "code-collapsed"}
digraph G {
    A -> B -> C
}
```

</details>

## Image collapsed mode

```dot render{"mode": "image-collapsed"}
digraph G {
    A -> B -> C
}
```

<details><summary>Image</summary>

![render-48fb4eec0b5d357512d90010cdef8bff.svg](render-48fb4eec0b5d357512d90010cdef8bff.svg)

</details>

## Code hidden mode

The code block will be hidden in this mode. It will still be viewable in in the
HTML source code as a HTML comment. The render directive for this code block is
`render{"mode": "code-hidden"}`.

![render-48fb4eec0b5d357512d90010cdef8bff.svg](render-48fb4eec0b5d357512d90010cdef8bff.svg)

<!--
```dot render{"mode": "code-hidden"}
digraph G {
    A -> B -> C
}
```
-->
