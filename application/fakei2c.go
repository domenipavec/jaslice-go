package application

type fakeI2C struct{}

func (f *fakeI2C) ReadByte(addr byte) (value byte, err error)               { return 0, nil }
func (f *fakeI2C) ReadBytes(addr byte, num int) (value []byte, err error)   { return nil, nil }
func (f *fakeI2C) WriteByte(addr, value byte) error                         { return nil }
func (f *fakeI2C) WriteBytes(addr byte, value []byte) error                 { return nil }
func (f *fakeI2C) ReadFromReg(addr, reg byte, value []byte) error           { return nil }
func (f *fakeI2C) ReadByteFromReg(addr, reg byte) (value byte, err error)   { return 0, nil }
func (f *fakeI2C) ReadWordFromReg(addr, reg byte) (value uint16, err error) { return 0, nil }
func (f *fakeI2C) WriteToReg(addr, reg byte, value []byte) error            { return nil }
func (f *fakeI2C) WriteByteToReg(addr, reg, value byte) error               { return nil }
func (f *fakeI2C) WriteWordToReg(addr, reg byte, value uint16) error        { return nil }
func (f *fakeI2C) Close() error                                             { return nil }
