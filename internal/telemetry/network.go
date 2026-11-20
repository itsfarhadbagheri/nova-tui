package telemetry

type InterfaceCounters struct {
	Name      string
	BytesSent uint64
	BytesRecv uint64
}

type NetworkCounters struct {
	Interfaces []InterfaceCounters
}

type NetworkRates struct {
	ReceivedBytesPerSecond float64
	SentBytesPerSecond     float64
	HasReceivedRate        bool
	HasSentRate            bool
}
