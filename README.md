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

## Features

- PlantUML and Graphviz diagrams
- SVG and PNG rendering
- Various output templates: `normal`, `code-collapsed`, `image-collapsed`, `code-hidden`
- Custom output filenames
- Images will only be re-rendered if the code block content has changed

## Usage

To mark a code block for rendering, add the `render` keyword to the opening fence.

    ```dot render
    digraph G {
        rankdir=LR;
        A -> B -> C;
    }
    ```

By default, the image will be rendered and placed above the code block.

    ![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

    ```dot render
    digraph G {
        rankdir=LR;
        A -> B -> C;
    }
    ```

The `render` keyword supports options, which can be specified in the form
`render{"optionName": "value"}`. Supported options are:

- `mode`: The placement of rendered images. Supported modes: `normal`
  (default), `code-collapsed`, `image-collapsed`, `code-hidden`.
- `filename`: The filename of the rendered image. If not specified, the
  filename will be automatically generated as `render-{hash}.svg`.

## Examples

I recommend viewing the [raw
source](https://github.com/benjaminheng/md-code-renderer/blob/master/README.md?plain=1)
of this readme as well.

### `normal` mode

Image is placed above the code block.

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

```dot render
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

### `code-collapsed` mode

Image is placed above the code block. Code block is collapsed.

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<details><summary>Source</summary>

```dot render{"mode": "code-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

</details>

### `image-collapsed` mode

Code block is placed above the image. Image is collapsed.

```dot render{"mode": "image-collapsed"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

<details><summary>Image</summary>

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

</details>

### `code-hidden` mode

The code block will be hidden in this mode. It will still be viewable in the
HTML source code as a HTML comment.

![render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg](./example/render-32455c4fc3bf7fc9a6c67d15f4cfd869.svg)

<!--
```dot render{"mode": "code-hidden"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```
-->

### Custom filename

The options for this code block is: `{"filename":
"readme-example-custom-filename.svg"}`. The generated image will use this
filename.

```dot render{"mode": "image-collapsed", "filename": "readme-example-custom-filename.svg"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

<details><summary>Image</summary>

![readme-example-custom-filename.svg](./example/readme-example-custom-filename.svg) <!-- hash:32455c4f -->

</details>

### Different output formats

If filename is specified, the output format is inferred from the file's
extension. In this example the filename has a `.png` extension, so a PNG image
is rendered.

```dot render{"mode": "image-collapsed", "filename": "readme-example-output-format-png.png"}
digraph G {
    rankdir=LR;
    A -> B -> C;
}
```

<details><summary>Image</summary>

![readme-example-output-format-png.png](./example/readme-example-output-format-png.png) <!-- hash:32455c4f -->

</details>
