# md-code-render

Renders code blocks in Markdown files into images, and inlines the images in the file.

- Supported languages: `dot`
- On the roadmap: `plantuml`

This is an experimental program for use in my knowledge base. The goal is to have code blocks containing diagramming DSLs, and be able to render them into inline images. There are a few existing implementations, but none of them do what I want. Most _replace_ the code block, whereas I want to inline the rendered image _alongside_ the code block. This tool is fairly opinionated at the moment. I may extend and generalize it in the future, though at the moment it does what I need it to.

## Example

```bash
md-code-render --types dot --output-dir ./static/diagrams/ --link-prefix "./diagrams/" ./file.md
```

Input:

    ```dot render
    digraph G {
        A -> B;
        B -> C -> D;
        B -> E;
    }
    ```

Output:

    ![render-34d8458eb066b9b3ac396fd5aee4c068.svg](/diagrams/render-34d8458eb066b9b3ac396fd5aee4c068.svg)

    ```dot render
    digraph G {
        A -> B;
        B -> C -> D;
        B -> E;
    }
    ```
