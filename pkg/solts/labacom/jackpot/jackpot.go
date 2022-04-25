package jackpot

type JackPot struct {
	Gold      int64
	SwingGold int64
	LastSwing int64
}

var Jp JackPot

func (j *JackPot) ChangeJacpot(v int64) {
	j.Gold += v
}

func (j *JackPot) Swing() {
	if j.Gold < 500*10000 {
		j.Gold += j.SwingGold
		j.LastSwing = j.SwingGold
	} else if j.Gold > 500*10000 {
		j.Gold -= j.SwingGold
		j.LastSwing = -j.SwingGold
	} else {
		j.Gold += j.LastSwing
	}
}

func (j *JackPot) GetJackpot() int64 {
	return j.Gold
}
