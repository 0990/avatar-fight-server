package util

func RandomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	ret := make([]byte, 0, length)
	for i := 0; i < length; i++ {
		ret = append(ret, bytes[Randn(len(bytes))])
	}
	return string(ret)
}
