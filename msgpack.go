package gouuidv6

// Msgpack Serialisation

func (u *UUID) ExtensionType() int8 { return 99 }

func (u *UUID) Len() int { return len(u) }

func (u *UUID) MarshalBinaryTo(b []byte) error {
	copy(b, (*u)[:])
	return nil
}
