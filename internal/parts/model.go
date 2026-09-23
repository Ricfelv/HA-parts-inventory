package parts

import "time"

var Types = []string{
	"Discrete_Capacitor",
	"Discrete_Crystal/Oscillator",
	"Discrete_Diode",
	"Discrete_Inductor",
	"Discrete_LED",
	"Discrete_MOSFET",
	"Discrete_Optocoupler",
	"Discrete_Resistor",
	"Discrete_Transistor",
	"Discrete_TVS/MOV",
	"IC_Amplifier",
	"IC_Driver",
	"IC_Logic",
	"IC_Memory",
	"IC_MCU",
	"IC_Other",
	"IC_Regulator",
	"IC_Transceiver",
	"Mech_Connector",
	"Mech_Fuse",
	"Mech_Motor",
	"Mech_Relay",
	"Mech_Speaker/Buzzer",
	"Mech_Switch",
	"Module_Connectivity",
	"Module_MCU",
	"Module_Power",
	"Module_Sensor",
	"Other",
	"Power_Battery",
	"Power_Supply",
	"Power_Transformer",
}

type Part struct {
	ID          int64
	Type        string
	Value       string
	Package     string
	Description string
	MPN         string
	Quantity    int
	UpdatedAt   time.Time
}

type ListOptions struct {
	Query       string
	InStockOnly bool
	Sort        string
	Direction   string
}

type Input struct {
	Type        string
	Value       string
	Package     string
	Description string
	MPN         string
	Quantity    int
}
