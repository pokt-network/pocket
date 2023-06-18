package runtime

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
	configTypes "github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var expectedGenesis = &genesis.GenesisState{
	GenesisTime: &timestamppb.Timestamp{
		Seconds: 1663610702,
		Nanos:   405401000,
	},
	ChainId:       "testnet",
	MaxBlockBytes: 4000000,
	Pools: []*types.Account{
		{
			Address: "44414f0000000000000000000000000000000000",
			Amount:  "100000000000000",
		},
		{
			Address: "466565436f6c6c6563746f720000000000000000",
			Amount:  "0",
		},
		{
			Address: "4170705374616b65506f6f6c0000000000000000",
			Amount:  "100000000000000",
		},
		{
			Address: "56616c696461746f725374616b65506f6f6c0000",
			Amount:  "100000000000000",
		},
		{
			Address: "53657276696365725374616b65506f6f6c000000",
			Amount:  "100000000000000",
		},
		{
			Address: "4669736865726d616e5374616b65506f6f6c0000",
			Amount:  "100000000000000",
		},
	},
	Accounts: []*types.Account{
		{
			Address: "00101f2ff54811e84df2d767c661f57a06349b7e",
			Amount:  "100000000000000",
		},
		{
			Address: "00201fa516d71eeff5c930d5b8ff99c23a948d25",
			Amount:  "100000000000000",
		},
		{
			Address: "00301e4a7012d207dcd777e8b5136fa4a6597970",
			Amount:  "100000000000000",
		},
		{
			Address: "004011e472a656063dfdf1ffb99cea74ffc8012f",
			Amount:  "100000000000000",
		},
		{
			Address: "005016a52b410047c5550ac9ebcb7db47cdf9718",
			Amount:  "100000000000000",
		},
		{
			Address: "0060177b5c13e00ddaaf3bfbb484b54a8231d4ba",
			Amount:  "100000000000000",
		},
		{
			Address: "00701ac438961045c465d159a2877f928e6bad97",
			Amount:  "100000000000000",
		},
		{
			Address: "0080144fc27b4c2392b642972758170f150246c2",
			Amount:  "100000000000000",
		},
		{
			Address: "0090176da8727b9a6fcf21596e78ec5a182dbc55",
			Amount:  "100000000000000",
		},
		{
			Address: "01001512e777bd8418439604ca27826b4228dcd2",
			Amount:  "100000000000000",
		},
		{
			Address: "01101e2cc48b08ce77469d8a8d97265778481601",
			Amount:  "100000000000000",
		},
		{
			Address: "01201712b60399e434689b09932d3051519e5ba3",
			Amount:  "100000000000000",
		},
		{
			Address: "01301fde2df9af256c9ee937863cc6a18145f3ba",
			Amount:  "100000000000000",
		},
		{
			Address: "0140199836364e4706cbfe0a4a423abb54b39701",
			Amount:  "100000000000000",
		},
		{
			Address: "015012c2e4ca5251d12b177ca4d3ff587debbdc3",
			Amount:  "100000000000000",
		},
		{
			Address: "016017165b4e37437533594b032a5822876715e7",
			Amount:  "100000000000000",
		},
		{
			Address: "017014190dee8fdeb53d91b0965c2a04ab212ac2",
			Amount:  "100000000000000",
		},
		{
			Address: "01801a9c4d1ad4bb6b382dc508526983fd8bc812",
			Amount:  "100000000000000",
		},
		{
			Address: "019016106aad11495ab6d8b0c2e46bde177301d5",
			Amount:  "100000000000000",
		},
		{
			Address: "0200147f81281e9f440f17db9ff1a75626e1eed9",
			Amount:  "100000000000000",
		},
		{
			Address: "021015ea03e16e366b2b6ea44e64c867a83b4b85",
			Amount:  "100000000000000",
		},
		{
			Address: "0220190659b54efd59135d00bb7156f35b0c3ff4",
			Amount:  "100000000000000",
		},
		{
			Address: "023011112dd9e662567b7d3a55f4e5a567bf680e",
			Amount:  "100000000000000",
		},
		{
			Address: "02401f68acd009b86dca7ceb39ca9f566bc503b2",
			Amount:  "100000000000000",
		},
		{
			Address: "02501641e0945c0b07183d7d1e6efdbd05ce0333",
			Amount:  "100000000000000",
		},
		{
			Address: "02601e36ee6667f91ea77e35b5ac8a177a70475f",
			Amount:  "100000000000000",
		},
		{
			Address: "027010db061a7c8f62c1355ce773b80301c5c316",
			Amount:  "100000000000000",
		},
		{
			Address: "0280113e2aa9069e2de98429fc70e14737be101d",
			Amount:  "100000000000000",
		},
		{
			Address: "0290159d3b4c6fbec3923572378f3acf94da84ee",
			Amount:  "100000000000000",
		},
		{
			Address: "030012adfa5ec4e125722df8053b91d8c5ad5a3a",
			Amount:  "100000000000000",
		},
		{
			Address: "03101fa9dfab749c26fa51308afe379d21697a0d",
			Amount:  "100000000000000",
		},
		{
			Address: "03201752aeff88f85bdd8c11b1b5c612368f143a",
			Amount:  "100000000000000",
		},
		{
			Address: "033019cc9c7ed32883f8f6674e9a0389a132e922",
			Amount:  "100000000000000",
		},
		{
			Address: "034014107398d1aed1849f87e0645c76722f6075",
			Amount:  "100000000000000",
		},
		{
			Address: "035014792bfec06cf9987e271443bd5166c6c875",
			Amount:  "100000000000000",
		},
		{
			Address: "03601e20421080533b3ccd5901e21002a0cc9b9d",
			Amount:  "100000000000000",
		},
		{
			Address: "03701140af54637215364552e9c9a6cdf99fe324",
			Amount:  "100000000000000",
		},
		{
			Address: "038018a16b79c80c25030286f01cba573966fe26",
			Amount:  "100000000000000",
		},
		{
			Address: "03901a2b80cab2a4891a630ce8d72bfd41a302f5",
			Amount:  "100000000000000",
		},
		{
			Address: "04001485986a1091505b8ca0599bf091d6206ad6",
			Amount:  "100000000000000",
		},
		{
			Address: "041018f6322a5be98f60056e4e2c417b7ad596da",
			Amount:  "100000000000000",
		},
		{
			Address: "0420119bec85a49ca58a6d1bcc7799cfd2ca38e5",
			Amount:  "100000000000000",
		},
		{
			Address: "0430139da13d5a19321ad5084741b5aeae55673e",
			Amount:  "100000000000000",
		},
		{
			Address: "04401470bc8cb00475736ec6d47467fab3f99b7d",
			Amount:  "100000000000000",
		},
		{
			Address: "0450188c4c9a39b3b944beedd1fd8e1424fae4aa",
			Amount:  "100000000000000",
		},
		{
			Address: "04601cb8c3f47d72841398459cdee4bed9d6dce0",
			Amount:  "100000000000000",
		},
		{
			Address: "04701414f3381cf05155ee0c49b46aa830e24f37",
			Amount:  "100000000000000",
		},
		{
			Address: "04801d1ba1d2fd2f52c73c0279c5757a69b8487a",
			Amount:  "100000000000000",
		},
		{
			Address: "04901a0a8fc0c2a1242ee43e947f46a8621ff0fc",
			Amount:  "100000000000000",
		},
		{
			Address: "05001dd51beced54ce86443ccbb313e077bd3d18",
			Amount:  "100000000000000",
		},
		{
			Address: "05101bcbe3058b23735b67a069f61ca0cd0879b3",
			Amount:  "100000000000000",
		},
		{
			Address: "05201b7902c59f14d85b74ee96f9bcc1b3fa60f3",
			Amount:  "100000000000000",
		},
		{
			Address: "05301e6da5753a57f130b0a6c53fbd3da4183b85",
			Amount:  "100000000000000",
		},
		{
			Address: "05401714dd94e63c10b0d07193c7c44eaac0a5ec",
			Amount:  "100000000000000",
		},
		{
			Address: "055010bbdc6f2bae265bbfd9b1e0b903e693b8c4",
			Amount:  "100000000000000",
		},
		{
			Address: "05601a412364210ccb99bdf5e8a512a054765972",
			Amount:  "100000000000000",
		},
		{
			Address: "057013122fe52e682b4e8de48804ca2bf5dffffe",
			Amount:  "100000000000000",
		},
		{
			Address: "05801cb9e2b56534210d580936ff5a680885136e",
			Amount:  "100000000000000",
		},
		{
			Address: "0590191280d4e4bffce2a33a7f0ac9a73bc9eb62",
			Amount:  "100000000000000",
		},
		{
			Address: "06001a00694e0b63f0f43505958a640dfcfcc8f6",
			Amount:  "100000000000000",
		},
		{
			Address: "0610106cda33303d6e00d7b5b794b86d57ba61a7",
			Amount:  "100000000000000",
		},
		{
			Address: "0620115bedf3acd970c9a59068d2b14adb7d857d",
			Amount:  "100000000000000",
		},
		{
			Address: "06301170a3835405154c7eb4f87094b41653681d",
			Amount:  "100000000000000",
		},
		{
			Address: "06401d158199047b5ef12afbcdce8db597aecded",
			Amount:  "100000000000000",
		},
		{
			Address: "06501d3f9f07c52b10d1e00e7ad1e4aae1a4a7e7",
			Amount:  "100000000000000",
		},
		{
			Address: "0660130363aba2aafe1c0622f93a130376a2be34",
			Amount:  "100000000000000",
		},
		{
			Address: "06701894ab795df825d427149d6a6ae78956d077",
			Amount:  "100000000000000",
		},
		{
			Address: "068011769a7d622cd0910dd14180e7c50d896d7b",
			Amount:  "100000000000000",
		},
		{
			Address: "06901695f154cb9b742bfda55abb8947e1385108",
			Amount:  "100000000000000",
		},
		{
			Address: "07001a8673a8e33385658f72dfd658426f9d734e",
			Amount:  "100000000000000",
		},
		{
			Address: "07101617e24df4bb1851c5d9e5e2a5ef0a03dd79",
			Amount:  "100000000000000",
		},
		{
			Address: "07201de9f4147025000a12d08b707f46aaa243d9",
			Amount:  "100000000000000",
		},
		{
			Address: "07301c797d17133923f5915b707e6bce48979e6f",
			Amount:  "100000000000000",
		},
		{
			Address: "07401c63472b26bcf6874d60beb35fd09eb0c7cc",
			Amount:  "100000000000000",
		},
		{
			Address: "07501b34052a99f7796ce85cd90fbfdf12182b6d",
			Amount:  "100000000000000",
		},
		{
			Address: "07601cfee9d4aaa549721acf6d2de9e985660f95",
			Amount:  "100000000000000",
		},
		{
			Address: "07701c2c51f22cdb8373fc1bb4ff9c01bf1487ba",
			Amount:  "100000000000000",
		},
		{
			Address: "07801bcc4c3f102c9632f924bfacb060971991f8",
			Amount:  "100000000000000",
		},
		{
			Address: "07901436353125d5154ebffd5578b91801b9aa4a",
			Amount:  "100000000000000",
		},
		{
			Address: "08001aea5d7ce40dda3f6118113e07fcb61e740b",
			Amount:  "100000000000000",
		},
		{
			Address: "081017a9f7df924e0e8c21ef9b5c4a30335ab1c1",
			Amount:  "100000000000000",
		},
		{
			Address: "0820168d7bd0858f1f75d502fc7b289f1ab7661a",
			Amount:  "100000000000000",
		},
		{
			Address: "08301572a443d79a6881b16771f3e5a064d62fac",
			Amount:  "100000000000000",
		},
		{
			Address: "084019507b6eb84225c85b3b41e3c6478eefe442",
			Amount:  "100000000000000",
		},
		{
			Address: "08501993c3184a796a8df29fb4d0e7105504a076",
			Amount:  "100000000000000",
		},
		{
			Address: "08601cfb0455379b0ab16d5ac5510d0c78b46536",
			Amount:  "100000000000000",
		},
		{
			Address: "08701a5fee5211f3acc5d228fedcb2446703cb8b",
			Amount:  "100000000000000",
		},
		{
			Address: "088018a5f5a9bdc98c71bba1917390852b924d97",
			Amount:  "100000000000000",
		},
		{
			Address: "08901a25edd020575b7c5b4a536ac2171183a495",
			Amount:  "100000000000000",
		},
		{
			Address: "09001a154cd1a4134dab17e0b4b9179950e7edc2",
			Amount:  "100000000000000",
		},
		{
			Address: "09101714d50bd08ae7600c348a91c93e2dc667b1",
			Amount:  "100000000000000",
		},
		{
			Address: "09201686ab546a441f2b0e1f14e4fa14636d094f",
			Amount:  "100000000000000",
		},
		{
			Address: "093014810acea8623624132a631f1415066b7acb",
			Amount:  "100000000000000",
		},
		{
			Address: "094016bb4ce71d3d3c6b3e41ab2da21ff9f4fce2",
			Amount:  "100000000000000",
		},
		{
			Address: "09501eaf63a0e625719b9a45ba47c6f88e9191b4",
			Amount:  "100000000000000",
		},
		{
			Address: "09601b86033e04c8b5bf061001bfcad00b76a29b",
			Amount:  "100000000000000",
		},
		{
			Address: "097013e595021019fd9f87383d273acbf1a1ee55",
			Amount:  "100000000000000",
		},
		{
			Address: "09801babadb26cfb04bc607e359083b497bfcf7f",
			Amount:  "100000000000000",
		},
		{
			Address: "099017b7a6845c52c1241b71ebe6a361e72eb770",
			Amount:  "100000000000000",
		},
		{
			Address: "001022b138896c4c5466ac86b24a9bbe249905c2",
			Amount:  "100000000000000",
		},
		{
			Address: "00202cd8f828a3818da2d24356984120f1cc3e8e",
			Amount:  "100000000000000",
		},
		{
			Address: "003028eec465dbbbbc0cdc62a6a5267d4987a9b9",
			Amount:  "100000000000000",
		},
		{
			Address: "0040221cede3b4cb812ada1b233252e29793bf8e",
			Amount:  "100000000000000",
		},
		{
			Address: "005027afddc9164b92f96f7ade10d736edf558f9",
			Amount:  "100000000000000",
		},
		{
			Address: "00602641744bb2ffb8f202e50e09339d3afa2f79",
			Amount:  "100000000000000",
		},
		{
			Address: "00702a91f6a10a74c4940b124ae689fc4005a274",
			Amount:  "100000000000000",
		},
		{
			Address: "008029c62d40726fcc05b141958d38401c238c9f",
			Amount:  "100000000000000",
		},
		{
			Address: "00902ceb992ec2e640aabd992dea518c92ff8421",
			Amount:  "100000000000000",
		},
		{
			Address: "01002e5bc517fc9a305da2eb03702f00cf209841",
			Amount:  "100000000000000",
		},
		{
			Address: "0110228bc2a06c366ac9d580ae8598cac0b19d07",
			Amount:  "100000000000000",
		},
		{
			Address: "0120275800be94c051cd7dceb8a9f9379e082cc4",
			Amount:  "100000000000000",
		},
		{
			Address: "01302f0a513476c2d86b7b1c1fa99e58c93579d6",
			Amount:  "100000000000000",
		},
		{
			Address: "014023b688126b9f70617bd9287e0128cbdb437d",
			Amount:  "100000000000000",
		},
		{
			Address: "01502136ffc149b1e2e809813bf3b69611d98bf1",
			Amount:  "100000000000000",
		},
		{
			Address: "01602db32e7be5da8e20974e1c9cb37fbb245f00",
			Amount:  "100000000000000",
		},
		{
			Address: "017028b6a1f045e0353910bf366df43aa7b7d387",
			Amount:  "100000000000000",
		},
		{
			Address: "01802cdbb497485fd1c302da02ff78007e8fde8f",
			Amount:  "100000000000000",
		},
		{
			Address: "01902f2d2353c18c119d8b9666c1e983d0ef1e98",
			Amount:  "100000000000000",
		},
		{
			Address: "020021a5e840756dfe23bfca87fd3ac85501a446",
			Amount:  "100000000000000",
		},
		{
			Address: "021020dccee8281c26d2da5ad9bb4f7ebd19da81",
			Amount:  "100000000000000",
		},
		{
			Address: "0220254d5d026614b128a9e886c6d64a047012af",
			Amount:  "100000000000000",
		},
		{
			Address: "02302ab455803ccc3f633dad3f29faa798d924f8",
			Amount:  "100000000000000",
		},
		{
			Address: "02402c706696618e43df284f6f6e76ae1e45bd55",
			Amount:  "100000000000000",
		},
		{
			Address: "02502e091a592e2bd77c185f7243ce76984eab17",
			Amount:  "100000000000000",
		},
		{
			Address: "0260255e6bb9d066f89d28b6fad1c15d3d89032f",
			Amount:  "100000000000000",
		},
		{
			Address: "027022745471f688cbfcc1ce5792479f76f46198",
			Amount:  "100000000000000",
		},
		{
			Address: "028021a4d053c3382cc21b3ac68379c9005d1b5c",
			Amount:  "100000000000000",
		},
		{
			Address: "02902688abadc84ff890d60f26f83b55e819a310",
			Amount:  "100000000000000",
		},
		{
			Address: "03002e1ccd1b3bf162928fc53ebd6deb43819c65",
			Amount:  "100000000000000",
		},
		{
			Address: "03102bb18cf59e6bfa0c332fa5d4a37f2b27c087",
			Amount:  "100000000000000",
		},
		{
			Address: "03202be2f7fe28b9d74f06b9255fe9b494e972dd",
			Amount:  "100000000000000",
		},
		{
			Address: "0330200c4f153e6f4c92458db19ada57c70ab3ef",
			Amount:  "100000000000000",
		},
		{
			Address: "03402db26738f220a362fe146ecf153642ff87da",
			Amount:  "100000000000000",
		},
		{
			Address: "03502b4d06a86695066222860a62d08df9ed2ab8",
			Amount:  "100000000000000",
		},
		{
			Address: "036028a16da00f47c7feb05065054269c69f9842",
			Amount:  "100000000000000",
		},
		{
			Address: "03702f52dcc8a56adfb6e3e1c7b3854b5c41284e",
			Amount:  "100000000000000",
		},
		{
			Address: "03802436e42318841ce50b3edd378b86db9d4752",
			Amount:  "100000000000000",
		},
		{
			Address: "03902ba2a29316e00d14b46a8dfa9e626adfe3a3",
			Amount:  "100000000000000",
		},
		{
			Address: "04002475796ef36b774db61e24b0849b0d51767c",
			Amount:  "100000000000000",
		},
		{
			Address: "04102b7dbe5ef655216fae8e78642b811f6dd442",
			Amount:  "100000000000000",
		},
		{
			Address: "04202d00ae70cc45eada50f78d9a6ecff7a5bd67",
			Amount:  "100000000000000",
		},
		{
			Address: "04302f0edec3bc299dd0fd5db7c2b01292129b1b",
			Amount:  "100000000000000",
		},
		{
			Address: "04402dee78188ecf8c9529d79faf81541adf4411",
			Amount:  "100000000000000",
		},
		{
			Address: "04502b0e45ad14ef276daa9170318226f6ea80bb",
			Amount:  "100000000000000",
		},
		{
			Address: "04602c471c5551fd710aa7688f10f786cce64570",
			Amount:  "100000000000000",
		},
		{
			Address: "0470271e21785451b9afc2113e25401a53d9b1a5",
			Amount:  "100000000000000",
		},
		{
			Address: "04802219a3d8193cf2dd2e02ce68145f417fe749",
			Amount:  "100000000000000",
		},
		{
			Address: "049022b51280be18f29bbdd2ee818f26c477f26f",
			Amount:  "100000000000000",
		},
		{
			Address: "050028893393d628df9cf4f1c169813b95b9b1cd",
			Amount:  "100000000000000",
		},
		{
			Address: "05102b721f1e19ece39df967d34f841586e020ba",
			Amount:  "100000000000000",
		},
		{
			Address: "05202d76ae73f5ee272101f81646cb80b3959b57",
			Amount:  "100000000000000",
		},
		{
			Address: "053022790202822ee0272599118b52a46faaf660",
			Amount:  "100000000000000",
		},
		{
			Address: "05402c5a5d082f48cde6dade4ca7ced08a49824b",
			Amount:  "100000000000000",
		},
		{
			Address: "05502b3731982758fe5d3dd3c39498409b5c9fd9",
			Amount:  "100000000000000",
		},
		{
			Address: "056024ebfbf3dad61c4d90ac7d3a46a2cbce988d",
			Amount:  "100000000000000",
		},
		{
			Address: "05702373f716c5dbb889e1d00ed7106dc4c1b055",
			Amount:  "100000000000000",
		},
		{
			Address: "0580262d9850e7030e4696cb5a8877c71c482f82",
			Amount:  "100000000000000",
		},
		{
			Address: "05902a94fcf4cb1269cf08ab8132555ee7255976",
			Amount:  "100000000000000",
		},
		{
			Address: "06002463ac77eeb9de8b8e08419d0ab186f348f3",
			Amount:  "100000000000000",
		},
		{
			Address: "06102038cd8e41c0e2c663e4c054601df492467b",
			Amount:  "100000000000000",
		},
		{
			Address: "062024ceb46015bfef52d8b345e1f8f0a0015d79",
			Amount:  "100000000000000",
		},
		{
			Address: "0630216fdf33e0e5b7b2927d6cf31f924136a992",
			Amount:  "100000000000000",
		},
		{
			Address: "06402d5245364cac56dc8aaed9f6f73bb7758edd",
			Amount:  "100000000000000",
		},
		{
			Address: "06502a00134512a1757399d52e5513c3db9ccddf",
			Amount:  "100000000000000",
		},
		{
			Address: "06602a848204191259e24c7b3ab77c8d289a9dd5",
			Amount:  "100000000000000",
		},
		{
			Address: "06702972e58b5c87ea9cd4ffea489b581b6dd7e9",
			Amount:  "100000000000000",
		},
		{
			Address: "06802c567f89d28db4f87f02cad5f13044e702a1",
			Amount:  "100000000000000",
		},
		{
			Address: "069021e9cab37247704722ff354370b7267711af",
			Amount:  "100000000000000",
		},
		{
			Address: "07002ce6f277af1eea10d238e5538c868301c763",
			Amount:  "100000000000000",
		},
		{
			Address: "07102b85a11230f14f61ae34d25ab36bf6828c6a",
			Amount:  "100000000000000",
		},
		{
			Address: "07202e3938b1d1643b0632e82e272a68a4c6cf4f",
			Amount:  "100000000000000",
		},
		{
			Address: "07302da7086df1e04bf4a9d59b3efbbd8d146891",
			Amount:  "100000000000000",
		},
		{
			Address: "0740250b7b3f77abe6d265d1ec95ba0fbc46f9e5",
			Amount:  "100000000000000",
		},
		{
			Address: "0750281e29134fd57731b0fd71288383da97fa30",
			Amount:  "100000000000000",
		},
		{
			Address: "07602b5a75c274ee5ba37473c3df6a943ebdaaf5",
			Amount:  "100000000000000",
		},
		{
			Address: "07702e8e221c50c9fc8d0a5a5ae245b9c021421e",
			Amount:  "100000000000000",
		},
		{
			Address: "07802c410ae33881ae12a69c0b465fa0d67f964d",
			Amount:  "100000000000000",
		},
		{
			Address: "079026d5277b4d4933e57464908f997bae3cab1a",
			Amount:  "100000000000000",
		},
		{
			Address: "08002547a37cb7cb15959612ed08dbc829740d6c",
			Amount:  "100000000000000",
		},
		{
			Address: "0810207041ae75b14e2beba59f3c1295fc8e3378",
			Amount:  "100000000000000",
		},
		{
			Address: "082026d1333477eac920f6e43b21bd8eb34c230c",
			Amount:  "100000000000000",
		},
		{
			Address: "08302c5e1331fb2970796e645bd4b6a74837b7da",
			Amount:  "100000000000000",
		},
		{
			Address: "08402876ffdf868a6749f495a000b6c594e888af",
			Amount:  "100000000000000",
		},
		{
			Address: "085022f50da0ad0d3c0842acc3ad07fdcf6146cd",
			Amount:  "100000000000000",
		},
		{
			Address: "086024cf8c5a24a989e14189f0f66c47f443f047",
			Amount:  "100000000000000",
		},
		{
			Address: "0870285683f485cf08cb75eb9e6282c6117bd9be",
			Amount:  "100000000000000",
		},
		{
			Address: "0880291617a3d4ea3f9209337c70c563fd6e71c5",
			Amount:  "100000000000000",
		},
		{
			Address: "0890281f0f414c6a8abc4b3f11779d5f81957cbe",
			Amount:  "100000000000000",
		},
		{
			Address: "09002fd2e2fca30f4895ffaaf67f864737ef4535",
			Amount:  "100000000000000",
		},
		{
			Address: "0910241b9a8642bd7db55bdb8b405c067872d153",
			Amount:  "100000000000000",
		},
		{
			Address: "092027d66154d48ee95fbc5dd786ec79d54d697b",
			Amount:  "100000000000000",
		},
		{
			Address: "09302c2e5f3e3effa561b9995b8af56c06185f93",
			Amount:  "100000000000000",
		},
		{
			Address: "09402c00c124b9a2dd763ecd123d4486efdcde94",
			Amount:  "100000000000000",
		},
		{
			Address: "0950282eed7a34205d122c606ec67966697eb04a",
			Amount:  "100000000000000",
		},
		{
			Address: "09602c9a9c8c906db1c95d81334b3e92c393cd93",
			Amount:  "100000000000000",
		},
		{
			Address: "097022d4a8563188d201bb50cbdf7defd1fba4d1",
			Amount:  "100000000000000",
		},
		{
			Address: "09802067f23d6361a42d4a6f48821b36bff7ae1b",
			Amount:  "100000000000000",
		},
		{
			Address: "0990276558e43b35ddf179fd2f45229c92a65153",
			Amount:  "100000000000000",
		},
		{
			Address: "0010336c3a2cc1ec71fecc45c360214f757194aa",
			Amount:  "100000000000000",
		},
		{
			Address: "00203f6f0cb678d38301b955103847b761dcf7c8",
			Amount:  "100000000000000",
		},
		{
			Address: "0030380ad041353c15cbb5a882dc46c76bd2d7a6",
			Amount:  "100000000000000",
		},
		{
			Address: "0040399ebce44befd43f9c21c4338b8099b0f020",
			Amount:  "100000000000000",
		},
		{
			Address: "005038e4fc6d542391f09999885a4ce97dba2997",
			Amount:  "100000000000000",
		},
		{
			Address: "00603b7a9c285cb334f3c5f7bcab812dcf579848",
			Amount:  "100000000000000",
		},
		{
			Address: "00703cbebfcb3a283305605146832b24f24f984b",
			Amount:  "100000000000000",
		},
		{
			Address: "00803983a54c2ce0001645628fe27d3b36fdbe56",
			Amount:  "100000000000000",
		},
		{
			Address: "00903263d1e8df5ec640ae3b8acf9dbfe32342ea",
			Amount:  "100000000000000",
		},
		{
			Address: "01003e823e60eba52ee8c4caa8f90cfdf0200a65",
			Amount:  "100000000000000",
		},
		{
			Address: "01103a454f68bb7ba6de23ca031f02788d453e3d",
			Amount:  "100000000000000",
		},
		{
			Address: "012038bee71c7845e3127f723a0aa31dc3746aaa",
			Amount:  "100000000000000",
		},
		{
			Address: "01303d938e55d4dea95b8100e5a57c9e1df0ab98",
			Amount:  "100000000000000",
		},
		{
			Address: "01403408c630df8e5d33093ae540058cfefddef6",
			Amount:  "100000000000000",
		},
		{
			Address: "01503d77b0a1a5f9ed1c863a3d351e79576878a5",
			Amount:  "100000000000000",
		},
		{
			Address: "01603ca8655e68453724f40ba4ed50f0e4259c24",
			Amount:  "100000000000000",
		},
		{
			Address: "0170385acad71c1689e969e31a9ff58e4ee972b8",
			Amount:  "100000000000000",
		},
		{
			Address: "01803615337cb54d124f45b73a12bc88a561b30e",
			Amount:  "100000000000000",
		},
		{
			Address: "01903d59c661d1ed66105a65fe4b1c2f2d1abe9e",
			Amount:  "100000000000000",
		},
		{
			Address: "020031a2be7b702f70f46bf829b6da14f55d9ae2",
			Amount:  "100000000000000",
		},
		{
			Address: "021036f7aa590ab4a07c2e8e96c0584ee88f952b",
			Amount:  "100000000000000",
		},
		{
			Address: "022035b7c1907874386dfa09c830b4018fd3cef3",
			Amount:  "100000000000000",
		},
		{
			Address: "02303818cb432bbccabec2f6300418353125871a",
			Amount:  "100000000000000",
		},
		{
			Address: "0240387ce52d9d5b7cbc3c98d4b5fa43682978ac",
			Amount:  "100000000000000",
		},
		{
			Address: "02503cf051e4c413cd816a8f48ca7f25a1c83336",
			Amount:  "100000000000000",
		},
		{
			Address: "026038347dc559de2d3126f805cc30505ea4acc2",
			Amount:  "100000000000000",
		},
		{
			Address: "02703e9b8d0a094af51f80e1d2ceaae68914a9eb",
			Amount:  "100000000000000",
		},
		{
			Address: "02803f00ae21b89a2200460b0c108f5851590d09",
			Amount:  "100000000000000",
		},
		{
			Address: "029037e23cfd8f607b8a3d44517f1d7b041b6b21",
			Amount:  "100000000000000",
		},
		{
			Address: "030036e7d54ab41985572e4bf907ee0ad9828db2",
			Amount:  "100000000000000",
		},
		{
			Address: "03103c3e6b426082a3dbf921e2a1f81da4bc8b19",
			Amount:  "100000000000000",
		},
		{
			Address: "032039cf359b2c307dbd842f8b508adc1b4a6a71",
			Amount:  "100000000000000",
		},
		{
			Address: "033039239d8a0d047cec1cf482b0b2376f8cc159",
			Amount:  "100000000000000",
		},
		{
			Address: "0340320b20f08a43fffd6925f46a6dcde8f7add5",
			Amount:  "100000000000000",
		},
		{
			Address: "03503e306772e56e4ea17a8d5bb7ee7a6b9e49c6",
			Amount:  "100000000000000",
		},
		{
			Address: "03603d967084e0b46d5c77d4a9b841419be9bbb5",
			Amount:  "100000000000000",
		},
		{
			Address: "03703a87580b213c1e7e579d106c75a7542dd39b",
			Amount:  "100000000000000",
		},
		{
			Address: "03803be0641fffaf5075fa7fd292792cc0f3daab",
			Amount:  "100000000000000",
		},
		{
			Address: "0390354912e5ef3487511f821118dcd64bd33694",
			Amount:  "100000000000000",
		},
		{
			Address: "0400323e23d6f9fc11376b67ad7161e797345e99",
			Amount:  "100000000000000",
		},
		{
			Address: "04103d7a0e6e01d4a433af271d6a9795dcf8650d",
			Amount:  "100000000000000",
		},
		{
			Address: "04203af178632271a1c4efde5f3885796bcf7928",
			Amount:  "100000000000000",
		},
		{
			Address: "04303ebbbf65ee90c30ef7c447adf9a668d0feea",
			Amount:  "100000000000000",
		},
		{
			Address: "044036a037784397e3d7e57446391de23a3ba4b3",
			Amount:  "100000000000000",
		},
		{
			Address: "045034ef28769401ae5b53a45b0134f331eba9c9",
			Amount:  "100000000000000",
		},
		{
			Address: "04603ad2b8acb92d580f7ad48783f6830792cc11",
			Amount:  "100000000000000",
		},
		{
			Address: "04703fea0e9c4a56b9d0d6f6b1a6e0bb011a2043",
			Amount:  "100000000000000",
		},
		{
			Address: "0480312a5d84c7d2748ae8fb79864ab8e9de44f0",
			Amount:  "100000000000000",
		},
		{
			Address: "04903b0390307695c0fdbd622a7418f89b3816d1",
			Amount:  "100000000000000",
		},
		{
			Address: "050034676cd28ddd59965a627ad26c19c87f8f42",
			Amount:  "100000000000000",
		},
		{
			Address: "051033bfb84a5283650fb6adffe4030978b796db",
			Amount:  "100000000000000",
		},
		{
			Address: "05203e18d18ce51ac2e36ec68938d90f0b2fe1fa",
			Amount:  "100000000000000",
		},
		{
			Address: "05303a5685c2654c1c065a37fedaca0d12e32931",
			Amount:  "100000000000000",
		},
		{
			Address: "05403bad71be2b49d41f3ffe14d0bbad77f030d3",
			Amount:  "100000000000000",
		},
		{
			Address: "05503afeb87fba4ff815ad7b1f9195320807da75",
			Amount:  "100000000000000",
		},
		{
			Address: "05603a40608f16f6dd85164182246959c6a822a4",
			Amount:  "100000000000000",
		},
		{
			Address: "05703d386e79008a5f769b1e4993d3e8322d7d41",
			Amount:  "100000000000000",
		},
		{
			Address: "0580330ebe42c355a020d43e515d45361239ab6f",
			Amount:  "100000000000000",
		},
		{
			Address: "059031cbc0127c2535276a88d559a6a4e4a53f2d",
			Amount:  "100000000000000",
		},
		{
			Address: "0600377f46ecc3ee1c80e832edb64e4dde33facf",
			Amount:  "100000000000000",
		},
		{
			Address: "061030d0ac519c2a3984f1d4cf94b6df3ec69669",
			Amount:  "100000000000000",
		},
		{
			Address: "06203a7ebb2ec52ae04af2027244714c3d3ddc07",
			Amount:  "100000000000000",
		},
		{
			Address: "06303274b07abed1b46c112b0a452061b64d3fd3",
			Amount:  "100000000000000",
		},
		{
			Address: "06403d9d3c8f49acb06bb90f35e6df2791a01860",
			Amount:  "100000000000000",
		},
		{
			Address: "065034257a38d4dbf0d3846a135134d5b621a605",
			Amount:  "100000000000000",
		},
		{
			Address: "066036d7ad4fe520d97eb1abbdcf4e49831be8b4",
			Amount:  "100000000000000",
		},
		{
			Address: "06703b2b859bb007eb509246f1d97cd24b9b9495",
			Amount:  "100000000000000",
		},
		{
			Address: "06803250816cac47b51cd719dbd1a81776d46f92",
			Amount:  "100000000000000",
		},
		{
			Address: "06903e6ca040c871d357f4dc9e6623a5f1b9b9c0",
			Amount:  "100000000000000",
		},
		{
			Address: "07003ffb728741b9ed161799bf98fce6c0315773",
			Amount:  "100000000000000",
		},
		{
			Address: "071037acb39111dab8ffe73a6f62cdff31b8051e",
			Amount:  "100000000000000",
		},
		{
			Address: "07203f1535bcc648e87d9ddd52dc8f0bea3f8d2c",
			Amount:  "100000000000000",
		},
		{
			Address: "07303339a9d4592b3f001c8e7c71179531514c90",
			Amount:  "100000000000000",
		},
		{
			Address: "0740387dff646af0cb748e2da550fbcfd4339f2f",
			Amount:  "100000000000000",
		},
		{
			Address: "07503df697c7519366ee104cd257b504ae215b69",
			Amount:  "100000000000000",
		},
		{
			Address: "0760387c0721331adc846186b1af29cc40eb101e",
			Amount:  "100000000000000",
		},
		{
			Address: "0770321a4900f0cdd73d0bb41a069efd0b00a05f",
			Amount:  "100000000000000",
		},
		{
			Address: "078038a968557ef3cfc4c1a441e84c2e7f9ef8a3",
			Amount:  "100000000000000",
		},
		{
			Address: "07903573d46e538999fc1cf759558d246ad88e27",
			Amount:  "100000000000000",
		},
		{
			Address: "080039d83b009515b0a889322a538e024ed61d29",
			Amount:  "100000000000000",
		},
		{
			Address: "08103713551e5be1acf40869b17aeb404d0aa008",
			Amount:  "100000000000000",
		},
		{
			Address: "082032d1136e29a08bbcba5b31fff6f6a571e65a",
			Amount:  "100000000000000",
		},
		{
			Address: "08303c6a8f0f40ae76b680e95b7cc6ea6c99cd4b",
			Amount:  "100000000000000",
		},
		{
			Address: "084031e989009c997cdac9f59d8bbcde2d2ba29d",
			Amount:  "100000000000000",
		},
		{
			Address: "08503a7301b4c7b4c137249afc4b9776eae58192",
			Amount:  "100000000000000",
		},
		{
			Address: "086038e00ea12c966f2cf2d37128783a046faa2a",
			Amount:  "100000000000000",
		},
		{
			Address: "08703e5420113fee0d82f6c0dd6510d0ea2a62d5",
			Amount:  "100000000000000",
		},
		{
			Address: "08803a6a4ff385eaf958f36b2e643648a3960dd8",
			Amount:  "100000000000000",
		},
		{
			Address: "0890315e9e60868ef016f657fd44522e0c4b5c35",
			Amount:  "100000000000000",
		},
		{
			Address: "090034972db5d1d1d75a7119efd07b19e5eae5c4",
			Amount:  "100000000000000",
		},
		{
			Address: "09103af9de0a7198d0ac38a5498213a8166d5967",
			Amount:  "100000000000000",
		},
		{
			Address: "09203ce3e0670d9345d9df3ab0b0bf05f3cb3db3",
			Amount:  "100000000000000",
		},
		{
			Address: "093030a0995e63ef34d120e1529649f0cfdcb1ac",
			Amount:  "100000000000000",
		},
		{
			Address: "09403ee296b05766a1cddb8cf9bc538e30b0a549",
			Amount:  "100000000000000",
		},
		{
			Address: "095034cf96d8fd373c041b322603cfc9701ecdbf",
			Amount:  "100000000000000",
		},
		{
			Address: "096035179a53391f4b8ff2af7fbc537bf1ac19bf",
			Amount:  "100000000000000",
		},
		{
			Address: "09703d73d215283cc75d053498484217c7054b3e",
			Amount:  "100000000000000",
		},
		{
			Address: "09803cf749f921c642ad2fe1052d700c46ae7807",
			Amount:  "100000000000000",
		},
		{
			Address: "09903f21cec8dbb97f9cb91fc62a764ad5a76ca1",
			Amount:  "100000000000000",
		},
		{
			Address: "00104055c00bed7c983a48aac7dc6335d7c607a7",
			Amount:  "100000000000000",
		},
		{
			Address: "00204737d2a165ebe4be3a7d5b0af905b0ea91d8",
			Amount:  "100000000000000",
		},
		{
			Address: "00304d0101847b37fd62e7bebfbdddecdbb7133e",
			Amount:  "100000000000000",
		},
		{
			Address: "00404a570febd061274f72b50d0a37f611dfe339",
			Amount:  "100000000000000",
		},
		{
			Address: "00504987d4b181c1e97b1da9af42f3db733b1ff4",
			Amount:  "100000000000000",
		},
		{
			Address: "00604d18001a2012830b93efcc23100450e5a512",
			Amount:  "100000000000000",
		},
		{
			Address: "007046b27ad5c49969d585dd3f4f7d1e4b94a42f",
			Amount:  "100000000000000",
		},
		{
			Address: "00804475b3c75e0fb86c88d937c7753f8cab8ba8",
			Amount:  "100000000000000",
		},
		{
			Address: "0090433ae78b5539c564134809502f699c90b009",
			Amount:  "100000000000000",
		},
		{
			Address: "0100471d6796652244885d1943ea7c13e3fce90b",
			Amount:  "100000000000000",
		},
		{
			Address: "011044141a86efafeae5ecc2c6290f0894072ab7",
			Amount:  "100000000000000",
		},
		{
			Address: "01204ed35af2bbef5d4db0aeedf3a9a8958c3296",
			Amount:  "100000000000000",
		},
		{
			Address: "0130404d6a39bf7cdc14ef4be4a0910f170edd1e",
			Amount:  "100000000000000",
		},
		{
			Address: "0140481810379d6e970f5b7de3220b70a784efd5",
			Amount:  "100000000000000",
		},
		{
			Address: "01504aedb93b489402d5aa654a5c6bda01880a5f",
			Amount:  "100000000000000",
		},
		{
			Address: "0160401ddec40b7a2dbc8505e4a9e998c22604b0",
			Amount:  "100000000000000",
		},
		{
			Address: "0170495866953b63cc9ba01c141dd8be029773bb",
			Amount:  "100000000000000",
		},
		{
			Address: "01804d0db5be444aeb1a114469a4ff7eb09c91db",
			Amount:  "100000000000000",
		},
		{
			Address: "01904e5651efaafb1a557d8f61034a4edfe0fa05",
			Amount:  "100000000000000",
		},
		{
			Address: "02004e81a4ee21929a399504fd0d946c545e7a2c",
			Amount:  "100000000000000",
		},
		{
			Address: "0210494fbd4b214dd6209d33b3fc960660c9f82b",
			Amount:  "100000000000000",
		},
		{
			Address: "0220406b71d941b5b9fc73a53d08ecb09c866b10",
			Amount:  "100000000000000",
		},
		{
			Address: "023041ae1ee33ad8db492951c13393b9bb59d59b",
			Amount:  "100000000000000",
		},
		{
			Address: "024041957fc3ac22552239b75d384f141f48430b",
			Amount:  "100000000000000",
		},
		{
			Address: "025045f51c4c956e2d822e6fd3f706957a5649fe",
			Amount:  "100000000000000",
		},
		{
			Address: "02604fcd21020ff140f6da784012e68d7f72ff36",
			Amount:  "100000000000000",
		},
		{
			Address: "0270460c5775e8951e5d43b3b37dc310209afda6",
			Amount:  "100000000000000",
		},
		{
			Address: "02804033a62cf08a2074cbffe941bd019576d5ed",
			Amount:  "100000000000000",
		},
		{
			Address: "0290479bfdf5210b7541b37b551716662a2ff397",
			Amount:  "100000000000000",
		},
		{
			Address: "03004511b4ca2d30b7e8bc008a4dc53588115fb0",
			Amount:  "100000000000000",
		},
		{
			Address: "031048ce3cef36af346cffc464dc5611028cc46f",
			Amount:  "100000000000000",
		},
		{
			Address: "03204ac71039cd18d1a23b76042c5c43c4dc1053",
			Amount:  "100000000000000",
		},
		{
			Address: "03304961130710e8984124354aa55eef028ad7d9",
			Amount:  "100000000000000",
		},
		{
			Address: "03404334ac8195996d072c916805dd26d82d9b0a",
			Amount:  "100000000000000",
		},
		{
			Address: "03504f415b35148ef268a37a7bdd5cbe41c46b62",
			Amount:  "100000000000000",
		},
		{
			Address: "0360444233608b747b82c9a651526171e0cecb09",
			Amount:  "100000000000000",
		},
		{
			Address: "037041d2ebb680f8e617c1d8bd4397114ed91d7a",
			Amount:  "100000000000000",
		},
		{
			Address: "03804e19cca9f5d8d074d3f0a60f1096723dad31",
			Amount:  "100000000000000",
		},
		{
			Address: "039044952cacd6fe9eed091c8b5b85c6e758d9ff",
			Amount:  "100000000000000",
		},
		{
			Address: "04004e7265815d837bac198381b1b3e70275319e",
			Amount:  "100000000000000",
		},
		{
			Address: "04104e28743787af30fdc85123b2ec763aca309c",
			Amount:  "100000000000000",
		},
		{
			Address: "042047946bbe90534f6940b3fa2aef34fa8f68c4",
			Amount:  "100000000000000",
		},
		{
			Address: "04304a7f8c6f59737bf2f6955745677351ad65bd",
			Amount:  "100000000000000",
		},
		{
			Address: "04404de47d3b8b90f533da541a6b9cd4b162cbdb",
			Amount:  "100000000000000",
		},
		{
			Address: "04504f18acbfe31c7f671a9986a499e74cd83b87",
			Amount:  "100000000000000",
		},
		{
			Address: "046047e21a2b35c94b8e25b39a8e11e088bdcc20",
			Amount:  "100000000000000",
		},
		{
			Address: "04704eff8a68e2d520bc9d4781a1328beb4a07fd",
			Amount:  "100000000000000",
		},
		{
			Address: "04804319d93a5aae6d6932e6255afa52ff9b81ad",
			Amount:  "100000000000000",
		},
		{
			Address: "0490494bc3baea23832c354eae4472d5ca97302f",
			Amount:  "100000000000000",
		},
		{
			Address: "05004facba0b2a253bab657f428338918d081158",
			Amount:  "100000000000000",
		},
		{
			Address: "0510462d8d3fbc0a5690e68507a49809f70afac4",
			Amount:  "100000000000000",
		},
		{
			Address: "05204497fa5d1cd636ec08c1e557aad11356ffdb",
			Amount:  "100000000000000",
		},
		{
			Address: "05304bebf8de8dd3333067868402df0a855cfc26",
			Amount:  "100000000000000",
		},
		{
			Address: "054041a4f8f82ae6cdc43ae50c621cbc73a74cba",
			Amount:  "100000000000000",
		},
		{
			Address: "055047122515c7448c7eb18b249e0bd390fe017b",
			Amount:  "100000000000000",
		},
		{
			Address: "056047aa21bce308658b7ff6ed0d87356de06220",
			Amount:  "100000000000000",
		},
		{
			Address: "05704c960137bbb6c9421a77e03671a334211ec5",
			Amount:  "100000000000000",
		},
		{
			Address: "058045f9c98e9fd506e5025797b34302fbaf85e3",
			Amount:  "100000000000000",
		},
		{
			Address: "05904f94b885d875f0fe6af3185a5936518b40f5",
			Amount:  "100000000000000",
		},
		{
			Address: "06004ef2c3c4979e68257b4bff94471e570b1cab",
			Amount:  "100000000000000",
		},
		{
			Address: "0610473810cd99f9ef4e72613cf5e93fcf942b12",
			Amount:  "100000000000000",
		},
		{
			Address: "0620415cc31eef8492058687df20c86b6e228bb2",
			Amount:  "100000000000000",
		},
		{
			Address: "06304edd5eb38af2ddc8555e94678ac6ab7a0a5c",
			Amount:  "100000000000000",
		},
		{
			Address: "06404246818eada50a3335e9c7e7eb142ee6709b",
			Amount:  "100000000000000",
		},
		{
			Address: "06504c11e6b449679278529efebf00ed4c1098b2",
			Amount:  "100000000000000",
		},
		{
			Address: "06604cea653df828010889eaa56f4ff867a17881",
			Amount:  "100000000000000",
		},
		{
			Address: "067044f84aa72577dc88b9813ee081dd6235905c",
			Amount:  "100000000000000",
		},
		{
			Address: "06804ac8b703a6500ba70f36a725bc55ecbe4bbc",
			Amount:  "100000000000000",
		},
		{
			Address: "06904f896ddabf920e559954efcb468e85e351d5",
			Amount:  "100000000000000",
		},
		{
			Address: "070043cc01e70228d252b4df847c4724813f6176",
			Amount:  "100000000000000",
		},
		{
			Address: "07104b360ef5589bb8b49a7e2c3e2fa03363a059",
			Amount:  "100000000000000",
		},
		{
			Address: "072048359ff76f59da57342b93dd4cfdd42662c2",
			Amount:  "100000000000000",
		},
		{
			Address: "07304f2af4afd31973cfbd0cae5a74722ba4fd14",
			Amount:  "100000000000000",
		},
		{
			Address: "07404ac156eb80fe3092c9e39b2169c6a087194f",
			Amount:  "100000000000000",
		},
		{
			Address: "07504a01ea1839c1aef06641eb46379218dbcba9",
			Amount:  "100000000000000",
		},
		{
			Address: "0760494df0c9bf14a5a5b50f491adfcc3df6b7e6",
			Amount:  "100000000000000",
		},
		{
			Address: "07704a472cabf1433eba2d80a73204d5726a88fc",
			Amount:  "100000000000000",
		},
		{
			Address: "0780445f4375d422c87a155b6949625b227bfba2",
			Amount:  "100000000000000",
		},
		{
			Address: "07904050bb8647901a3e9b93a2296ca4334d728e",
			Amount:  "100000000000000",
		},
		{
			Address: "08004bdfd6dca205244f5a15f9d1741bcc511b54",
			Amount:  "100000000000000",
		},
		{
			Address: "081046c92e38c4202ff659168d45d742a57ffee2",
			Amount:  "100000000000000",
		},
		{
			Address: "082045bad7d45189d4c9fb379de5ae3aa7dff762",
			Amount:  "100000000000000",
		},
		{
			Address: "08304f805360247bed9730d07e755384cb0070e5",
			Amount:  "100000000000000",
		},
		{
			Address: "08404debeec7ee380d24cd384d96ff98912a450b",
			Amount:  "100000000000000",
		},
		{
			Address: "085047967a6d86f2369345e1243af970f7e4c5e2",
			Amount:  "100000000000000",
		},
		{
			Address: "086044b59a0f9b0ceee5ee9f298d17c9f1445beb",
			Amount:  "100000000000000",
		},
		{
			Address: "08704f20c81ebf51988fb1b6bef5d29c702216c9",
			Amount:  "100000000000000",
		},
		{
			Address: "08804f6d3823b4839685744cd807f5ff7f4d9712",
			Amount:  "100000000000000",
		},
		{
			Address: "089049f0d03a767dd1a2b9479cb70d2c836a0077",
			Amount:  "100000000000000",
		},
		{
			Address: "09004e53f819778cd2c0738ff055416f44baf529",
			Amount:  "100000000000000",
		},
		{
			Address: "09104c7006a42917b1d6c824dc48a4e1f87f835f",
			Amount:  "100000000000000",
		},
		{
			Address: "09204a46380169c1c9b5cbff225e771e17e32cda",
			Amount:  "100000000000000",
		},
		{
			Address: "093041155ee61c12ea790594b7b305b00ed179e6",
			Amount:  "100000000000000",
		},
		{
			Address: "09404741ec9ee6c40d009174474db97aee4db4b4",
			Amount:  "100000000000000",
		},
		{
			Address: "095049ee9b15b848f432b1fef028641e391de082",
			Amount:  "100000000000000",
		},
		{
			Address: "096046a684f4677dde3a3db075d3b3f99359d3aa",
			Amount:  "100000000000000",
		},
		{
			Address: "097048eba43973cc73cd10f1b18085424318d145",
			Amount:  "100000000000000",
		},
		{
			Address: "0980419ea31998de5d84b43f5042e4e492a5d738",
			Amount:  "100000000000000",
		},
		{
			Address: "099044a5282c4089976c6b13ab13e5423dae55ce",
			Amount:  "100000000000000",
		},
	},
	Applications: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_APP,
			Address:         "00101f2ff54811e84df2d767c661f57a06349b7e",
			PublicKey:       "bb851ac31120a4c8848738582f358599abbc3d84638f8fa79f74aeafad1eede0",
			Chains:          []string{"0001"},
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00101f2ff54811e84df2d767c661f57a06349b7e",
		},
	},
	Validators: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "00104055c00bed7c983a48aac7dc6335d7c607a7",
			PublicKey:       "dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
			Chains:          nil,
			ServiceUrl:      "validator1:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00104055c00bed7c983a48aac7dc6335d7c607a7",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "00204737d2a165ebe4be3a7d5b0af905b0ea91d8",
			PublicKey:       "eb2c78364525a210d994a83e02d18b4287ab81f6670cf4510ab6c9f51e296d91",
			Chains:          nil,
			ServiceUrl:      "validator2:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00204737d2a165ebe4be3a7d5b0af905b0ea91d8",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "00304d0101847b37fd62e7bebfbdddecdbb7133e",
			PublicKey:       "1041a9c76539791fef9bee5b4fcd5bf4a1a489e0790c44cbdfa776b901e13b50",
			Chains:          nil,
			ServiceUrl:      "validator3:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00304d0101847b37fd62e7bebfbdddecdbb7133e",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "00404a570febd061274f72b50d0a37f611dfe339",
			PublicKey:       "d6cea8706f6ee6672c1e013e667ec8c46231e0e7abcf97ba35d89fceb8edae45",
			Chains:          nil,
			ServiceUrl:      "validator4:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00404a570febd061274f72b50d0a37f611dfe339",
		},
	},
	Servicers: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_SERVICER,
			Address:         "00104055c00bed7c983a48aac7dc6335d7c607a7",
			PublicKey:       "dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
			Chains:          []string{"0001"},
			ServiceUrl:      "validator1:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "00104055c00bed7c983a48aac7dc6335d7c607a7",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_SERVICER,
			Address:         "001022b138896c4c5466ac86b24a9bbe249905c2",
			PublicKey:       "56915c1270bc8d9280a633e0be51647f62388a851318381614877ef2ed84a495",
			Chains:          []string{"0001"},
			ServiceUrl:      "servicer1:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "001022b138896c4c5466ac86b24a9bbe249905c2",
		},
	},
	Fishermen: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_FISH,
			Address:         "0010336c3a2cc1ec71fecc45c360214f757194aa",
			PublicKey:       "d913a05a6f4bde35413bdcc6343238960cfc7d8aff425fb712dcaa52f1476dbf",
			Chains:          []string{"0001"},
			ServiceUrl:      "fisherman1:42069",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "0010336c3a2cc1ec71fecc45c360214f757194aa",
		},
	},
	Params: test_artifacts.DefaultParams(),
}

func TestNewManagerFromReaders(t *testing.T) {
	type args struct {
		configReader  io.Reader
		genesisReader io.Reader
		options       []func(*Manager)
	}
	defaultCfg := configs.NewDefaultConfig()
	buildConfigBytes, err := os.ReadFile("../build/config/config.validator1.json")
	if err != nil {
		require.NoError(t, err)
	}

	buildGenesisBytes, err := os.ReadFile("../build/config/genesis.json")
	if err != nil {
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		args      args
		want      *Manager
		assertion require.ComparisonAssertionFunc
	}{
		{
			name: "reading from the build directory",
			args: args{
				configReader:  strings.NewReader(string(buildConfigBytes)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				config: &configs.Config{
					RootDirectory: "/go/src/github.com/pocket-network",
					PrivateKey:    "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
					NetworkId:     "localnet",
					Consensus: &configs.ConsensusConfig{
						PrivateKey:      "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
						MaxMempoolBytes: 500000000,
						PacemakerConfig: &configs.PacemakerConfig{
							TimeoutMsec:               10000,
							Manual:                    true,
							DebugTimeBetweenStepsMsec: 1000,
						},
						ServerModeEnabled: true,
					},
					Utility: &configs.UtilityConfig{
						MaxMempoolTransactionBytes: 1073741824,
						MaxMempoolTransactions:     9000,
					},
					Persistence: &configs.PersistenceConfig{
						PostgresUrl:       "postgres://postgres:postgres@pocket-db:5432/postgres",
						NodeSchema:        "validator1",
						BlockStorePath:    "/var/blockstore",
						TxIndexerPath:     "/var/txindexer",
						TreesStoreDir:     "/var/trees",
						MaxConnsCount:     50,
						MinConnsCount:     1,
						MaxConnLifetime:   "5m",
						MaxConnIdleTime:   "1m",
						HealthCheckPeriod: "30s",
					},
					P2P: &configs.P2PConfig{
						PrivateKey:     "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
						Hostname:       "validator1",
						Port:           defaults.DefaultP2PPort,
						ConnectionType: configTypes.ConnectionType_TCPConnection,
						MaxNonces:      1e5,
					},
					Telemetry: &configs.TelemetryConfig{
						Enabled:  true,
						Address:  "0.0.0.0:9000",
						Endpoint: "/metrics",
					},
					Logger: &configs.LoggerConfig{
						Level:  "debug",
						Format: "pretty",
					},
					RPC: &configs.RPCConfig{
						Enabled: true,
						Port:    "50832",
						Timeout: 30000,
						UseCors: false,
					},
					Keybase: defaultCfg.Keybase,
					Servicer: &configs.ServicerConfig{
						Enabled: true,
						Chains:  []string{"0001"},
					},
					Validator: &configs.ValidatorConfig{Enabled: true},
					Fisherman: defaultCfg.Fisherman,
					IBC:       &configs.IBCConfig{Enabled: true},
				},
				genesisState: expectedGenesis,
				clock:        clock.New(),
			},
			assertion: func(tt require.TestingT, want, got any, _ ...any) {
				require.Equal(tt, want.(*Manager).config, got.(*Manager).config)
				require.Equal(tt, want.(*Manager).genesisState, got.(*Manager).genesisState)
			},
		},
		{
			name: "unset MaxNonces should fallback to default value",
			args: args{
				configReader: strings.NewReader(string(`{
					"p2p": {
					  "hostname": "validator1",
					  "port": 42069,
					  "use_rain_tree": true,
					  "is_empty_connection_type": false,
					  "private_key": "4ff3292ff14213149446f8208942b35439cb4b2c5e819f41fb612e880b5614bdd6cea8706f6ee6672c1e013e667ec8c46231e0e7abcf97ba35d89fceb8edae45"
					}
				  }`)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				config: &configs.Config{
					P2P: &configs.P2PConfig{
						PrivateKey:     "4ff3292ff14213149446f8208942b35439cb4b2c5e819f41fb612e880b5614bdd6cea8706f6ee6672c1e013e667ec8c46231e0e7abcf97ba35d89fceb8edae45",
						Hostname:       "validator1",
						Port:           42069,
						ConnectionType: configTypes.ConnectionType_TCPConnection,
						MaxNonces:      defaults.DefaultP2PMaxNonces,
					},
					Keybase: defaultCfg.Keybase,
				},
				genesisState: expectedGenesis,
				clock:        clock.New(),
			},
			assertion: func(tt require.TestingT, want, got any, _ ...any) {
				require.Equal(tt, want.(*Manager).config.P2P, got.(*Manager).config.P2P)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewManagerFromReaders(tt.args.configReader, tt.args.genesisReader, tt.args.options...)
			tt.assertion(t, tt.want, got)
		})
	}
}
