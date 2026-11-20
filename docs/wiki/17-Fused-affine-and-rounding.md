# Fused affine transforms and rounding

An affine transform scales a value and adds an offset. Calibration and unit conversion often use this form.

## How the methods differ

With separate operations, float32 rounds the product before adding the offset and rounds again. A fused multiply-add rounds only after computing the product plus the offset. When large terms nearly cancel, the first rounding can make a visible difference to the residual.

For scale 1 + 2⁻²³, input 1 − 2⁻²³, and offset −1, the exact result is
−2⁻⁴⁶. Separate multiply and add round the product to 1 before subtracting 1,
so the result is positive zero. Fusion keeps −2⁻⁴⁶.

## What the benchmark measures

The timed work includes applying the transform to the whole batch, including broadcasts, loads, and stores. Software and vector FMA are different measuring conditions. Check the recorded path before attributing a difference to hardware.

## Try it

Try **Broadcast affine (unfused)** and **Broadcast affine (fused binary32)** with the same size and shared constants. For a numerical comparison, use the rounding-sensitive fixture described under [fixture options](03-Command-registry#fixture-options).

See also [three-array fused multiply-add](24-Three-array-fused-multiply-add).
