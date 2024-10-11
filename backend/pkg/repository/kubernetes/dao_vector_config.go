package kubernetes

func (k *k8sApi) GetVectorConfigFile() (map[string]string, error) {
	return k.getConfigMap(k.VectorCMName, k.VectorFileName)
}

func (k *k8sApi) UpdateVectorConfigFile(content []byte) error {
	return k.updateConfigMap(k.VectorCMName, k.VectorFileName, content)
}
