package gokakera

func (k *Kakera) ExportComputeChecksum(data []byte) (string, error) {
	return k.computeChecksum(data)
}
