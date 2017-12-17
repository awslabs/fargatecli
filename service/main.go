package service

import (
	"errors"
)

const mebibytesInGibibyte = 1024
const cpuUnitsInVCpu = 1024

var InvalidCpuAndMemoryCombination = errors.New(`Invalid CPU and Memory settings

CPU (CPU Units)    Memory (MiB)
---------------    ------------
256                512, 1024, or 2048
512                1024 through 4096 in 1GB increments
1024               2048 through 8192 in 1GB increments
2048               4096 through 16384 in 1GB increments
4096               8192 through 30720 in 1GB increments
`)

func ValidateCpuAndMemory(cpuUnits int16, mebibytes int16) error {
	switch cpuUnits {
	case 256:
		if mebibytes == 512 || validateMebibytes(mebibytes, 1024, 2048) {
			return nil
		}
	case 512:
		if validateMebibytes(mebibytes, 1024, 4096) {
			return nil
		}
	case 1024:
		if validateMebibytes(mebibytes, 2048, 8192) {
			return nil
		}
	case 2048:
		if validateMebibytes(mebibytes, 4096, 16384) {
			return nil
		}
	case 4096:
		if validateMebibytes(mebibytes, 8192, 30720) {
			return nil
		}
	}

	return InvalidCpuAndMemoryCombination
}

func validateMebibytes(mebibytes, min, max int16) bool {
	return mebibytes >= min && mebibytes <= max && mebibytes%mebibytesInGibibyte == 0
}
