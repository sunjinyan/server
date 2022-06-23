package id



type AccountId string

func (a AccountId)String()string  {
	return string(a)
}



type TripId string

func (a TripId)String()string  {
	return string(a)
}

type IdentityID string

func (i IdentityID)String()string  {
	return string(i)
}

type CarID string

func (c CarID)String()string  {
	return string(c)
}


type BlobID string

func (c BlobID)String()string  {
	return string(c)
}