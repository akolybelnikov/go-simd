# What is SIMD?

**SIMD** stands for **Single Instruction, Multiple Data**. It lets a CPU
apply the same operation to several values at once. A group of values is
called a **vector**, and each position in it is a **lane**.

## One example

Suppose you want to add two lists of numbers, pair by pair:

| Array | Lane 1 | Lane 2 | Lane 3 | Lane 4 |
| --- | ---: | ---: | ---: | ---: |
| First list | 1 | 2 | 3 | 4 |
| Second list | 10 | 20 | 30 | 40 |
| Result | 11 | 22 | 33 | 44 |

A **scalar** loop adds one pair at a time. With a four-lane vector, a CPU
can add all four pairs together. The program still has to read the inputs
and write the results. If there’s a fifth pair, it belongs to the **tail**,
the values left over after the full group. SIMD works within a CPU core,
and multiple cores can each use SIMD. [Jesús Espino’s SIMD and Go](https://internals-for-interns.com/posts/go-simd/)

## When it helps

SIMD suits repeated, independent work, such as adjusting image pixels,
scaling audio samples, or adding numeric columns. Short lists may spend
more time on setup and leftovers than they save on arithmetic. Scattered
data or decisions that differ for each value can make it harder to group
the work. Even when the arithmetic gets faster, reading and writing the
data can still take most of the time. [SIMD and Go](https://internals-for-interns.com/posts/go-simd/) · [LLVM’s vectorization guide](https://llvm.org/docs/Vectorizers.html)

When adding a whole list into one total, the vector method may combine the
numbers in a different order. That can change the rounded total.
[Sum and reduction order](14-Sum-and-reduction-order)

## SIMD in Go

Go 1.27 offers an experimental `simd` package. It lets Go programs describe
work in vectors and use the width available on the machine. Go can use a
software fallback when suitable hardware support is unavailable, so
portable code can have different timings on different machines. [Official Go SIMD overview](https://go.dev/blog/simd-experiment) · [Package documentation](https://pkg.go.dev/simd)

Try **Addition** in **Explore**. Compare ordinary and vectorized addition
at 1,024 values, then try a much smaller batch. In **Results**, look at the
median and the individual repeats. [Your first addition comparison](02-First-experiment)
