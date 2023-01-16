package tlsutils

type PemBlockType string

const (
	PemBlockTypeCertificate   PemBlockType = "CERTIFICATE"
	PemBlockTypeCsr                        = "CERTIFICATE REQUEST"
	PemBlockTypePrivateKey                 = "PRIVATE KEY"
	PemBlockTypeRSAPrivateKey              = "RSA PRIVATE KEY"
	PemBlockTypePublicKey                  = "PUBLIC KEY"
)