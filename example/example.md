# Example

## Normal mode

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

```dot render
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

```dot render{"mode": "normal"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

## Code collapsed mode

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<details><summary>Source</summary>

```dot render{"mode": "code-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

</details>

## Image collapsed mode

```dot render{"mode": "image-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

<details><summary>Image</summary>

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

</details>

## Code hidden mode

The code block will be hidden in this mode. It will still be viewable in in the
HTML source code as a HTML comment. The render directive for this code block is
`render{"mode": "code-hidden"}`.

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<!--
```dot render{"mode": "code-hidden"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```
-->
