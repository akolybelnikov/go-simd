# Glossary and capabilities

Use **Explore** to see the operations available on the measuring machine. The [measurement guide](32-Operation-measurement-guide) explains what each operation does, where it is used, and what the timing includes.

| Term | Meaning |
| --- | --- |
| Scalar | A method that works on individual input values in source order. |
| SIMD | Single instruction, multiple data. A vector operation expresses work on several lanes at once. |
| Lane | One value position in a vector. Lane count depends on vector width and the type of value. |
| Tail | The values left after processing the full vector chunks. |
| Fixture | A repeatable set of generated inputs chosen for the experiment. |
| Predicate | A comparison that yields a true or false decision. |
| Mask | A collection of decisions used to select values or stored for later processing. |
| Reduction | An operation that combines multiple values into one, such as a sum or dot product. |
| Fold order | The order in which a reduction combines its values. Rounding can make order matter. |
| Fused multiply-add | Computing a product plus an addend with one final rounding. |
| Median | The middle repeat, or the mean of the two middle repeats. |
| Logical bytes | The amount of input and output data associated with the operation. Physical memory traffic can differ. |
| Emulation | Running vector semantics through a software path. Check the recorded measuring conditions. |

The lab compares its own operations on generated inputs. A result applies to the measured batch, methods, and machine. See [Understanding results](06-Understanding-results) when deciding whether a comparison answers your question.
