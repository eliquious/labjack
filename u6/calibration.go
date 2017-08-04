package u6

// CalibrationInfo holds the U6 calibration
type CalibrationInfo struct {
	ProductID    uint8
	HiResolution bool
	CalConstants CalibrationConstants
}

// CalibrationConstants holds the calibration constants
type CalibrationConstants [40]float64

/*
   Calibration constants order
   0 - AIN +-10V Slope, GainIndex=0
   1 - AIN +-10V Offset, GainIndex=0
   2 - AIN +-1V Slope, GainIndex=1
   3 - AIN +-1V Offset, GainIndex=1
   4 - AIN +-100mV Slope, GainIndex=2
   5 - AIN +-100mV Offset, GainIndex=2
   6 - AIN +-10mV Slope, GainIndex=3
   7 - AIN +-10mV Offset, GainIndex=3
   8 - AIN +-10V Neg. Slope, GainIndex=0
   9 - AIN +-10V Center Pt., GainIndex=0
   10 - AIN +-1V Neg. Slope, GainIndex=1
   11 - AIN +-1V Center Pt., GainIndex=1
   12 - AIN +-100mV Neg. Slope, GainIndex=2
   13 - AIN +-100mV Center Pt., GainIndex=2
   14 - AIN +-10mV Neg. Slope, GainIndex=3
   15 - AIN +-10mV Center Pt., GainIndex=3
   16 - DAC0 Slope
   17 - DAC0 Offset
   18 - DAC1 Slope
   19 - DAC1 Offset
   20 - Current Output 0
   21 - Current Output 1
   22 - Temperature Slope
   23 - Temperature Offset

   High Resolution
   24 - AIN +-10V Slope, GainIndex=0
   25 - AIN +-10V Offset, GainIndex=0
   26 - AIN +-1V Slope, GainIndex=1
   27 - AIN +-1V Offset, GainIndex=1
   28 - AIN +-100mV Slope, GainIndex=2
   29 - AIN +-100mV Offset, GainIndex=2
   30 - AIN +-10mV Slope, GainIndex=3
   31 - AIN +-10mV Offset, GainIndex=3
   32 - AIN +-10V Neg. Slope, GainIndex=0
   33 - AIN +-10V Center Pt., GainIndex=0
   34 - AIN +-1V Neg. Slope, GainIndex=1
   35 - AIN +-1V Center Pt., GainIndex=1
   36 - AIN +-100mV Neg. Slope, GainIndex=2
   37 - AIN +-100mV Center Pt., GainIndex=2
   38 - AIN +-10mV Neg. Slope, GainIndex=3
   39 - AIN +-10mV Center Pt., GainIndex=3
*/

// DefaultCalibrationInfo holds the default values.
var DefaultCalibrationInfo = CalibrationInfo{
	ProductID:    6,
	HiResolution: false,
	CalConstants: [40]float64{
		0.00031580578,
		-10.5869565220,
		0.000031580578,
		-1.05869565220,
		0.0000031580578,
		-0.105869565220,
		0.00000031580578,
		-0.0105869565220,
		-.000315805800,
		33523.0,
		-.0000315805800,
		33523.0,
		-.00000315805800,
		33523.0,
		-.000000315805800,
		33523.0,
		13200.0,
		0.0,
		13200.0,
		0.0,
		0.00001,
		0.0002,
		-92.379,
		465.129,
		0.00031580578,
		-10.5869565220,
		0.000031580578,
		-1.05869565220,
		0.0000031580578,
		-0.105869565220,
		0.00000031580578,
		-0.0105869565220,
		-.000315805800,
		33523.0,
		-.0000315805800,
		33523.0,
		-.00000315805800,
		33523.0,
		-.000000315805800,
		33523.0,
	},
}
