# md-code-renderer

Renders code blocks in Markdown files into images, and inlines the images in the file.

- Supported languages: `dot`, `plantuml`

This is an experimental program for use in my knowledge base. The goal is to
have code blocks containing diagramming DSLs, and be able to render them into
inline images. There are a few existing implementations, but none of them do
what I want. Most _replace_ the code block, whereas I want to inline the
rendered image _alongside_ the code block. This tool is fairly opinionated at
the moment. I may extend and generalize it in the future, though at the moment
it does what I need it to.

## Example

```bash
md-code-renderer render --languages dot --output-dir ./static/diagrams/ --link-prefix "./diagrams/" ./file.md
```

To mark a code block for rendering, add the `render` keyword to the opening fence.

    ```dot render
    digraph G {
        A -> B;
        B -> C -> D;
        B -> E;
    }
    ```

By default, the image will be rendered and placed above the code block.

    ![render-34d8458eb066b9b3ac396fd5aee4c068.svg](/diagrams/render-34d8458eb066b9b3ac396fd5aee4c068.svg)

    ```dot render
    digraph G {
        A -> B;
        B -> C -> D;
        B -> E;
    }
    ```

The `render` keyword supports options, which can be specified in the form
`render{"mode": "normal"}`. If no options are passed, the default mode is
`normal`.

Supported modes are:

- `normal`
- `code-collapsed`
- `image-collapsed`
- `code-hidden`

`normal` mode:

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

```dot render
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

`code-collapsed` mode:

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<details><summary>Source</summary>

```dot render{"mode": "code-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

</details>

`image-collapsed` mode:

```dot render{"mode": "image-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

<details><summary>Image</summary>

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

</details>

`code-hidden` mode:

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<!--
```dot render{"mode": "code-hidden"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```
-->
