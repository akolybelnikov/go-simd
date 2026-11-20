# Low64 carryless polynomial multiplication

Carryless multiplication treats bits as polynomial coefficients over GF(2). Such products are building blocks for checksums and finite-field arithmetic.

## How the methods differ

Coefficients add by XOR and never produce an integer carry. For example, the carryless product of 3 and 3 is 5. The scalar method shifts and XORs partial products. The vector method obtains the products for even and odd lanes and keeps each low word. Both discard product bits beyond 64.

## What the benchmark measures

The benchmark times reads, multiplication, lane extraction, temporary work, and output stores. A runtime capability flag tells you what was available on that host. It does not by itself identify the machine instructions emitted.

## Try it

Select **Low64 carryless polynomial multiplication** in **Explore**. Compare lengths near the measured uint64 lane count and keep the runtime capability fixed across runs.

See also [bitwise XOR](28-XOR-of-float32-bit-patterns).
