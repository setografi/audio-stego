package spectral

import "gonum.org/v1/gonum/fourier"

// ForwardTransform applies the Fast Fourier Transform (FFT) to the signal.
func ForwardTransform(signal []float64) []complex128 {
    // Convert signal to complex128
    complexSignal := make([]complex128, len(signal))
    for i, v := range signal {
        complexSignal[i] = complex(v, 0)
    }

    // Apply FFT
    fft := fourier.NewFFT(len(signal))
    fft.Transform(complexSignal)
    return complexSignal
}

// InverseTransform applies the Inverse Fast Fourier Transform (IFFT) to the frequency coefficients.
func InverseTransform(coeffs []complex128) []float64 {
    // Apply IFFT
    fft := fourier.NewFFT(len(coeffs))
    fft.Inverse(coeffs)

    // Convert complex128 to float64
    signal := make([]float64, len(coeffs))
    for i, v := range coeffs {
        signal[i] = real(v)
    }
    return signal
}
