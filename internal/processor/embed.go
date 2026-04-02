package processor

import _ "embed"

//go:embed processor.py
var processorPy []byte

//go:embed requirements.txt
var requirementsTxt []byte
