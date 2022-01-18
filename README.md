# md-code-render

Renders code blocks in Markdown files into images, and inlines the images in the file.

- Supported languages: `dot`
- On the roadmap: `plantuml`

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
