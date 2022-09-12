package utils

import (
	"testing"

	bscript2 "github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

var (
	p2pkHex     = "410444e56eab3d6f4aca5e71f51b3fe389951af2a030e14cc33dc8f665c5af28f65875898b4dc59ab1bb2071e625d8140b4fead2706fd43ad907339aaf0e090315dcac"
	p2pkhHex    = "76a91413473d21dc9e1fb392f05a028b447b165a052d4d88ac"
	p2shHex     = "a9149bc6f9caddaaab28c2bc0a8bf8531f91109bdd5887"
	metanetHex  = "006a046d65746142303237383763323464643466..."
	opReturnHex = "006a067477657463684d9501424945"
	multisigHex = "514104cc71eb30d653c0c3163990c47b976f3fb3f37cccdcbedb169a1dfef58bbfbfaff7d8a473e7e2e6d317b87bafe8bde97e3cf8f065dec022b51d11fcdd0d348ac4410461cbdcc5409fb4b4d42b51d33381354d80e550078cb532a34bfa2fcfdeb7d76519aecc62770f5b0e4ef8551946d8a540911abe3e7854a26f39f58b25c15342af52ae"
	stasHex     = "76a9146d3562a8ec96bcb3b2253fd34f38a556fb66733d88ac6976aa607f5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7c5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01007e818b21414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff007d976e7c5296a06394677768827601249301307c7e23022079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798027e7c7e7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01417e21038ff83d8cf12121491609c4939dc11c4aa35503508fe432dc5a5c1905608b9218ad547f7701207f01207f7701247f517f7801007e8102fd00a063546752687f7801007e817f727e7b01177f777b557a766471567a577a786354807e7e676d68aa880067765158a569765187645294567a5379587a7e7e78637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6867567a6876aa587a7d54807e577a597a5a7a786354807e6f7e7eaa727c7e676d6e7eaa7c687b7eaa587a7d877663516752687c72879b69537a647500687c7b547f77517f7853a0916901247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77788c6301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f777852946301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77686877517f7c52797d8b9f7c53a09b91697c76638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6876638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6863587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f7768587f517f7801007e817602fc00a06302fd00a063546752687f7801007e81727e7b7b687f75537f7c0376a9148801147f775379645579887567726881766968789263556753687a76026c057f7701147f8263517f7c766301007e817f7c6775006877686b537992635379528763547a6b547a6b677c6b567a6b537a7c717c71716868547a587f7c81547a557964936755795187637c686b687c547f7701207f75748c7a7669765880748c7a76567a876457790376a9147e7c7e557967041976a9147c7e0288ac687e7e5579636c766976748c7a9d58807e6c0376a9147e748c7a7e6c7e7e676c766b8263828c007c80517e846864745aa0637c748c7a76697d937b7b58807e56790376a9147e748c7a7e55797e7e6868686c567a5187637500678263828c007c80517e846868647459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e687459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e68687c537a9d547963557958807e041976a91455797e0288ac7e7e68aa87726d77776a14f566909f378788e61108d619e40df2757455d14c010005546f6b656e"
	stas2Hex    = "76a914e130e550626fb267992ea4180f9aaf04ed96357688ac6976aa607f5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7c5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01007e818b21414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff007d976e7c5296a06394677768827601249301307c7e23022079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798027e7c7e7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01417e21038ff83d8cf12121491609c4939dc11c4aa35503508fe432dc5a5c1905608b9218ad547f7701207f01207f7701247f517f7801007e8102fd00a063546752687f7801007e817f727e7b01177f777b557a766471567a577a786354807e7e676d68aa880067765158a569765187645294567a5379587a7e7e78637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6867567a6876aa587a7d54807e577a597a5a7a786354807e6f7e7eaa727c7e676d6e7eaa7c687b7eaa587a7d877663516752687c72879b69537a647500687c7b547f77517f7853a0916901247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77788c6301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f777852946301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77686877517f7c52797d8b9f7c53a09b91697c76638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6876638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6863587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f7768587f517f7801007e817602fc00a06302fd00a063546752687f7801007e81727e7b7b687f75537f7c0376a9148801147f775379645579887567726881766968789263556753687a76026c057f7701147f8263517f7c766301007e817f7c6775006877686b537992635379528763547a6b547a6b677c6b567a6b537a7c717c71716868547a587f7c81547a557964936755795187637c686b687c547f7701207f75748c7a7669765880748c7a76567a876457790376a9147e7c7e557967041976a9147c7e0288ac687e7e5579636c766976748c7a9d58807e6c0376a9147e748c7a7e6c7e7e676c766b8263828c007c80517e846864745aa0637c748c7a76697d937b7b58807e56790376a9147e748c7a7e55797e7e6868686c567a5187637500678263828c007c80517e846868647459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e687459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e68687c537a9d547963557958807e041976a91455797e0288ac7e7e68aa87726d77776a14f566909f378788e61108d619e40df2757455d14c010005546f6b656e"
	sensibleHex = "5101400176018801a901ac01240873656e7369626c65607601249376011493768b765493760124567993760114937601149376589376011493768b768b765a9376011493760280017614fa9c10692ebb28864635ad7c6911752ce6b43602145d15eedd93c90d91e0d76de5cc932c833baf833614777e4dd291059c9f7a0fd563f7204576dcceb79101195514e506495141d845848a851dbdd11022d255188506013c79762097dfd76851bf465e8f715593b217714858bbe9570ff3bd5e33840a34e20ff0262102ba79df5f8ae7604a9830f03c7933028186aede0675a16f025dc4f8be8eec0382210ac407f0e4bd44bfc207355a778b046225a7068fc59ee7eda43ad905aadbffc800206c266b30e6a1319c66dc401e5bd6b432ba49688eecd118297041da8074ce0810201008ce7480da41702918d1ec8e6849ba32b4d65b1e40dc669c31a1e6306b266c012679012679855679aa767d517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e01007e817757795679567956795679537956795479577995939521414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff006e6e977b757d009f636e937b757c68757b757c6e5296a063765279947b757c6853798277527982775452799378930130787e527e53797e57797e527e52797e5579517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7e56797e7777777777777777777777765779ac7777777777777777777769013b79aa013d797601247f75547f777788013b79827702d800a16901237900a069013c797601687f7700005279517f75007f7d7701fd87635379537f75517f7d7701007e8177537a757b7b5379535479937f75537f777b757c677601fe87635379557f75517f7d7701007e8177537a757b7b5379555479937f75557f777b757c677601ff87635379597f75517f7d7701007e8177537a757b7b5379595479937f75597f777b757c675379517f75007f7d7701007e8177537a757b7b5379515479937f75517f777b757c686868757777777682776e7f75780114947f776f750114947f7552790128947f77700128947f755379013c947f7701307901317982776e011e79940114937f7578011e79947f77777776013b79013f79013f79013f790145790145790145795679a95879884f53007600a26976539f69946b6c766b796c78775279a0697d7757007600a26976539f69946b6c766b796c750117796e8b80767682778c7f75007f77777777597952798b0114957f7552790114957f7778a9886d53517600a26976539f69946b6c766b796c78775279a0697d7757517600a26976539f69946b6c766b796c750117796e8b80767682778c7f75007f77777777597952798b0114957f7552790114957f7778a9886d53527600a26976539f69946b6c766b796c78775279a0697d7757527600a26976539f69946b6c766b796c750117796e8b80767682778c7f75007f77777777597952798b0114957f7552790114957f7778a988756d6d6d6d6d013079000000537901247f75537a757b7b5379012c7f7501247f7701007e817b757c5379012c7f77776f755279537a75537a75537a75537a75014079014079014079013679013679013679013679013679013679013679014e7952014079005d007600a26976539f69946b6c766b796c7557007600a26976539f69946b6c766b796c755b007600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d5d517600a26976539f69946b6c766b796c7557517600a26976539f69946b6c766b796c755b517600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d5d527600a26976539f69946b6c766b796c7557527600a26976539f69946b6c766b796c755b527600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d76539d5a79000000537901247f75537a757b7b5379012c7f7501247f7701007e817b757c5379012c7f77776f755279537a75537a75537a75537a75567956798b013579957f755679013579957f7753797888785679a98851776d6d6d6d6d6d6d6d6d0134795f79587952798277707c52796e012879940124937f7578012779947f77a977778878547952796e6e706e012b799454937f7578012b79947f7701007e817777760129799377777776780078014ba1635177677802ff00a1635277677803ffff00a16353776755776868687793705279947f7577777777a9777788537978000000547954796e6094012879947f7578011894012879947f7701007e817777537a757b7b547954796e5894012879947f75786094012879947f7701007e8177777b757c547954796e012879947f75785894012879947f7701007e817777776f755279537a75537a75537a75537a75537a756f755279537a75537a75537a75537a75537a75537a75537a75567901407901417982776e012179940114937f7578012179947f7777778801437901437901437901417901417901417901417901417901417901417901517953014b79005d007600a26976539f69946b6c766b796c7557007600a26976539f69946b6c766b796c755b007600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d5d517600a26976539f69946b6c766b796c7557517600a26976539f69946b6c766b796c755b517600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d5d527600a26976539f69946b6c766b796c7557527600a26976539f69946b6c766b796c755b527600a26976539f69946b6c766b796c755d79787e76a87676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e777676a87e7d7701007e8177775279537995547997785579979c6354798b557a75547a547a547a547a547975686d6d76539d5a79000000537901247f75537a757b7b5379012c7f7501247f7701007e817b757c5379012c7f77776f755279537a75537a75537a75537a75567956798b013879957f755679013879957f7753797888785679a98851776d6d6d6d6d6d6d6d6d607960795a7901427976827754796f756e012d79947f7578012579947f77a977778853796f756e6e0120799452947f75007f777777a977778852796f756e012379940114937f7578012379947f777777886e6e0123799458937f7578012379947f7d7701007e8177777777777777777653799d013079021027011179949576547995557901337993021027959601327901117995021027967602f4019f63007768567901347993789459790135799352799457795479947600a069013e79013f798277547953795b7954795479011894012c79947f755379586e8b80767682778c7f75007f777777777e5279586e8b80767682778c7f75007f777777777e78586e8b80767682778c7f75007f777777777e55795579012d79947f777e77777777777653797658805279768277007802fd009f6378516e8b80767682778c7f75007f77777777776778030000019f6301fd5279526e8b80767682778c7f75007f777777777e7767780500000000019f6301fe5279546e8b80767682778c7f75007f777777777e776778090000000000000000019f6301ff5279586e8b80767682778c7f75007f777777777e77686868687653797e7777777e77770148790149798277011279597970012879947f75007f7752797e78586e8b80767682778c7f75007f777777777e547954797f755479012b79947f777e77777777760139797658805279768277007802fd009f6378516e8b80767682778c7f75007f77777777776778030000019f6301fd5279526e8b80767682778c7f75007f777777777e7767780500000000019f6301fe5279546e8b80767682778c7f75007f777777777e776778090000000000000000019f6301ff5279586e8b80767682778c7f75007f777777777e77686868687653797e7777777e77775a795979014c79011679013d795479547994006ea0635479557982775579547970013179947f75007f7752797e78586e8b80767682778c7f75007f777777777e547954797f755479013479947f777e777777777654797658805279768277007802fd009f6378516e8b80767682778c7f75007f77777777776778030000019f6301fd5279526e8b80767682778c7f75007f777777777e7767780500000000019f6301fe5279546e8b80767682778c7f75007f777777777e776778090000000000000000019f6301ff5279586e8b80767682778c7f75007f777777777e77686868687653797e7777777e777777776877777777777701397901397900527900a063780139790138797e01147e787e0139797e0137797e777654797658805279768277007802fd009f6378516e8b80767682778c7f75007f77777777776778030000019f6301fd5279526e8b80767682778c7f75007f777777777e7767780500000000019f6301fe5279546e8b80767682778c7f75007f777777777e776778090000000000000000019f6301ff5279586e8b80767682778c7f75007f777777777e77686868687653797e7777777e777777776877775979011a7900527900a06378013a790139797e01147e787e013a797e0138797e777654797658805279768277007802fd009f6378516e8b80767682778c7f75007f77777777776778030000019f6301fd5279526e8b80767682778c7f75007f777777777e7767780500000000019f6301fe5279546e8b80767682778c7f75007f777777777e776778090000000000000000019f6301ff5279586e8b80767682778c7f75007f777777777e77686868687653797e7777777e77777777687777557954797e53797e52797e787e76aa0158797682776e58947f75780128947f77777787777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777777776a3c1ab05711510da1794d8f761124e10e55b183e21cdd76a35dca7fa1a5a2d480cfac11b76c3cc0ec7b6248e91bfa08cc1e0a2379b6fe35e2017c0dbbda"
)

// TestIsP2PK will test the method IsP2PK()
func TestIsP2PK(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsP2PK(p2pkHex))
	})

	t.Run("no match - extra data", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH(p2pkHex+"06"))
	})
}

// TestIsP2PKH will test the method IsP2PKH()
func TestIsP2PKH(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsP2PKH(p2pkhHex))
	})

	t.Run("no match - extra data", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH(p2pkhHex+"06"))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, P2PKHSubstringRegexp.MatchString("somethesetstring"+p2pkhHex+"06rtdhrth"))
	})
}

// TestIsP2SH will test the method IsP2SH()
func TestIsP2SH(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsP2SH("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsP2SH(p2shHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, P2SHSubstringRegexp.MatchString("test"+p2shHex+"test"))
	})
}

// TestIsMetanet will test the method IsOpReturn()
func TestIsMetanet(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsMetanet("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsMetanet(metanetHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, MetanetSubstringRegexp.MatchString("test"+metanetHex+"test"))
	})
}

// TestIsOpReturn will test the method IsOpReturn()
func TestIsOpReturn(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsOpReturn("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsOpReturn(opReturnHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, OpReturnSubstringRegexp.MatchString("test"+opReturnHex+"test"))
	})
}

// TestIsStas will test the method IsStas()
func TestIsStas(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsStas("nope"))
	})

	t.Run("no match - p2pkhHex", func(t *testing.T) {
		assert.Equal(t, false, IsStas(p2pkhHex))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsStas(stasHex))
	})

	t.Run("match 2", func(t *testing.T) {
		assert.Equal(t, true, IsStas(stas2Hex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, StasSubstringRegexp.MatchString("test"+stas2Hex+"test"))
	})
}

// TestIsSensible will test the method IsSensible()
func TestIsSensible(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsSensible("nope"))
	})

	t.Run("no match - p2pkhHex", func(t *testing.T) {
		assert.Equal(t, false, IsSensible(p2pkhHex))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsSensible(sensibleHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, SensibleSubstringRegexp.MatchString("test"+sensibleHex+"test"))
	})
}

// TestIsMultiSig will test the method IsMultiSig()
func TestIsMultiSig(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsMultiSig("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsMultiSig(multisigHex))
	})
}

// TestGetDestinationType will test the method GetDestinationType()
func TestGetDestinationType(t *testing.T) {
	t.Parallel()

	t.Run("no match - non standard", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeNonStandard, GetDestinationType("nope"))
	})

	t.Run("ScriptTypePubKey", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypePubKey, GetDestinationType(p2pkHex))
	})

	t.Run("ScriptTypePubKeyHash", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypePubKeyHash, GetDestinationType(p2pkhHex))
	})

	t.Run("ScriptHashType", func(t *testing.T) {
		assert.Equal(t, ScriptHashType, GetDestinationType(p2shHex))
	})

	t.Run("metanet - ScriptMetanet", func(t *testing.T) {
		assert.Equal(t, ScriptMetanet, GetDestinationType(metanetHex))
	})

	t.Run("op return - ScriptTypeNullData", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeNullData, GetDestinationType(opReturnHex))
	})

	t.Run("multisig - ScriptTypeMultiSig", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeMultiSig, GetDestinationType(multisigHex))
	})

	t.Run("stas - ScriptTypeTokenStas", func(t *testing.T) {
		assert.Equal(t, ScriptTypeTokenStas, GetDestinationType(stas2Hex))
	})
}

// TestGetAddressFromScript will test the method GetAddressFromScript()
func TestGetAddressFromScript(t *testing.T) {
	t.Parallel()

	t.Run("p2pk", func(t *testing.T) {
		assert.Equal(t, "1BYpPJHowiz9Qr6zsTzRXKNeej2RV2Av6H", GetAddressFromScript(p2pkHex))
	})

	t.Run("p2pkh", func(t *testing.T) {
		assert.Equal(t, "12kwBQPUnAMouxBBWRa5wsA6vC29soEdXT", GetAddressFromScript(p2pkhHex))
	})

	t.Run("stas 1", func(t *testing.T) {
		assert.Equal(t, "1AxScC72W9tyk1Enej6dBsVZNkkgAonk4H", GetAddressFromScript(stasHex))
	})

	t.Run("stas 2", func(t *testing.T) {
		assert.Equal(t, "1MXhcVvUz1LGSkoUFGkANHXkGCtrzFKHpA", GetAddressFromScript(stas2Hex))
	})

	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, "", GetAddressFromScript("invalid-or-unknown-script"))
	})
}

func BenchmarkIsP2PKH(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsP2PKH(p2pkhHex)
	}
}

func BenchmarkGetDestinationTypeRegex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetDestinationType(stas2Hex)
	}
}
