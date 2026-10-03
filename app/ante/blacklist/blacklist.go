package blacklist

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var Blacklist = map[string]bool{
	"terra1kvwkvurw4xexw69ef772p95jajnjjjsq6d5uca": true, // victim
	"terra1r0kyl9k9apsd5rp8qvjnda2prg76d4e6s8zymg": true, // victim
	"terra1x04xgtwlw72gtfzrq7nfwmr6eexla8ecljw28z": true, // victim
	"terra10pa0qv7mrqst053zhv9hxxt9mwhmr45ggguygs": true, // victim
	"terra10nxpnrcm78rlevvt28ztucwf6xcqts26yj4vfl": true, // victim
	"terra1tvshfczvr9lgp9aytwjdhclmgsdkr4gqnxugpe": true, // victim
	"terra1dt5deezcgv2mzxlhksqghv5x35yy3sapepl0t2": true, // victim
	"terra12v9ltjgct6qsaw28j5m26z7ruprhxywkmjfspk": true, // victim
	"terra1tz9gyy75ucws9fqfx282h3syvvy43lflg686qg": true, // victim
	"terra1a4xyusa48u9cx678cvx23jajh4ca6v8ec2pguu": true, // victim
	"terra16g7thcqtqkg6x42kej34r2kzhnl8ghvp2j4d4m": true, // victim
	"terra1akqq32nm6x545htxejp5dh5tvdtmk987tnnaqk": true, // victim
	"terra1g6054czmwvle0rahwazkdpmtllyfyhnvjr94hn": true, // victim
	"terra1760w9ckqt5xh4urs235vm27xxkdwgeunay06v3": true, // victim
	"terra188c33gkknt3njnrkzunfz0fudjfn3hrr4jgure": true, // victim
	"terra13huy6a34qj2mza2c2cntq95vg93u8x9aag5c6s": true, // victim
	"terra16gymuk00l9684ctz5dzqtgrgqywyf7hz8pfr87": true, // victim
	"terra1taevkzkdw3vw2cemhgxnvwc8yl5753pgsvxxz3": true, // victim
	"terra1rksuz09vqfgrax2ryyadsy2xrqwzj8j5d275am": true, // victim
	"terra1n79knz98spcfxt5g0hmc4vcmw8lhmw0v4nssyv": true, // victim
	"terra1a7lrn0h92vy9t3jxnll5ad022xaf5tcapt7azf": true, // victim
	"terra13y0vz2a9mygluaasmel6kafxmxlkneygp4344f": true, // victim
	"terra1g4a7fhwt6cft2myf3fne7vw5tfdkvyh95d3an6": true, // victim
	"terra19l4n2nn563zdg2rhqy0yjg3fm6hw22ntnzzmnk": true, // victim
	"terra1u7vg8crqf09xwjya6d5q4e66tfxe48arl69nz5": true, // victim
	"terra1q0nnd3zuz7aexzdmy8guv4kmc98mmjmhntjarx": true, // victim
	"terra17vw4q8pr8k22yuc9q7m8fmxud8s8dpkvncfpg8": true, // victim
	"terra1xpart34e6wmkhscyj58npm0xeerew77jgh7mc0": true, // victim
	"terra13zua9z2gtp0sl4tf236eykculdg6k8l7nj39nc": true, // victim
	"terra1kwawc69v6th5aucmj3g4r5njhkk6mgyrn8un9g": true, // victim
	"terra1fkgfwcd0an2cgk8yyh7jnx35waxapjzf75egd4": true, // victim
	"terra1r2k9v7regc3e94636e2rcxkejuf6dfyt4ulz7p": true, // victim
	"terra19cyjvyhtrvvqmwu830gf5xws0tvqapx92yvy26": true, // victim
	"terra1ucuecygdq28jvapdydj7awef4ka80xgzhx9dfv": true, // victim
	"terra1qvp7pv8k80yrw4kcywz2ujhxmu35aak8xnvr4a": true, // victim
	"terra1c7ume8f0aq7kdzl2yltpl25yn6cv6rkxaaq4pj": true, // victim
	"terra19k4kqjqwj530wtpmwwpfd2q6kl3v9tl2jp2dff": true, // victim
	"terra1zmrr3h8s8cptk3rtqxw45fxvvftqsxwh2kdy85": true, // victim
	"terra1k86y2pedv6hzhtr5d7pgy7frsc27n5quc6tqyq": true, // victim
	"terra1nv4zxl07s5cnja3aljs5a5xgczwcyrygt56dm5": true, // victim
	"terra1jefyjcyrurn83edjvjj8vqncv3mdv9pgrwnhkh": true, // victim
	"terra1j4va7q245mu9cs0ls8v7wgk4yvncj5cwn8cetp": true, // victim
	"terra1ycx5jee4f4tkkse45x8qhvnu8hmk4kt25d3p2v": true, // victim
	"terra1gpw0kxu45rl3uaj6qtv4k7xntazk89uvag49sp": true, // victim
	"terra1pc3vp2ua6rxyp0eycyhmc0zdsmyknf22r8tvds": true, // victim
	"terra136fcgj0xw6dgmf85w44tyzc560tytlr20sfnax": true, // victim
	"terra1fqgwjedr6wc63f5hasmurzc7kfqe6zqjz2ets7": true, // victim
	"terra1hke2efh926ees978cf229e2qra3yjcacrye8ay": true, // victim
	"terra1q332a4mn7v72hfkd6lfu5awhzmv7uxefhz4l2s": true, // victim
	"terra1mrzrpmdcxdw3ydzgpugxkca7ds7lzrl32dnyrr": true, // victim
	"terra1fzyp2j2zvtaapzmpxjklyy4ss7tc894qzl3dm3": true, // victim
	"terra17pl6mgedyy448ynvupkhz5yvfflqxszhtx9vja": true, // victim
	"terra1f2hswusq0zfwzj2mr9xvy443744mvkxxqyp6kj": true, // victim
	"terra19afgrt5c5xxphcx08p7qnvc4v764n07w46vwt7": true, // victim
	"terra1p908f42jr2us7yxdsmjfhcw033rl9awah5wfk3": true, // victim
	"terra1aamhc54nxyf0y7dmnu3ly8m83sdcj8pge645yd": true, // victim
	"terra1p0mglpkuv06a7xryrxlg707nlzhlj0pvr3trz7": true, // victim
	"terra1tyj09ulfgpr5m7xx0hmp3qk4qylld8dmcctkk8": true, // victim
	"terra15qglzdgl7598wkg353f5x8yag4q2xa0xdaqq32": true, // victim
	"terra1krvdnpgndl9e00w5vfs7avg9gff9c4qzj03e5t": true, // victim
	"terra1cpevlen8hnzg7ndhgywenqmkl2wpk08l58nlcs": true, // victim
	"terra1ld3nkdackd6nhlr3s3mp45e5zh52yk2fytcy9t": true, // victim
	"terra1mzcd2weugnce5fzdlkmqnjsrzv4hewa9vv2kqk": true, // victim
	"terra1lvw2kay4mj33zaej0m6grdqhu4yg98hujwdf26": true, // victim
	"terra132hzuh98s7utegmvdleznr0kvx9cspkqhzg8xs": true, // victim
	"terra1sgaastm2whl47m7kyepre9qu762qekj9sllpzy": true, // victim
	"terra1v5dqd8fdck68tlqveulqwctdctrcl59ydyrem2": true, // victim
	"terra1qd3nesm6weghwu9yhwq6wj85ux25e9w9k0f545": true, // victim
	"terra1kkcaw5h5c0mvy62fpzr0wry5mz6hjqppcfejpy": true, // victim
	"terra16tmh6k508x47l9kq437jyev33tn3t2mm2txqnf": true, // victim
	"terra1krk02jgfjd490tn8xewnd7erp935f94sejrwtc": true, // victim
	"terra1np5l2y7y2jm5sdl2hklyjx0wm4swhmyufwfju9": true, // victim
	"terra1yqk4e4e9mx6gyg094sfzq6n8tmzkz956kxc8td": true, // victim
	"terra1amghnkxp2yzyaenvs0s7dedr9seyfa8yaq7qtw": true, // victim
	"terra1hwyv80p8arn9ueh2209lmadgsyd6ypcvps2pnz": true, // victim
	"terra18tk7yvad335nku0632kvx8l8jdfqa9pk0q5yma": true, // victim
	"terra1r6lhgtt6kppx7us0e2wgejpqhqx4ke58ck4n2e": true, // victim
	"terra1gvj2jayujddtl9e8jewmlcg84mansljf4ztvrf": true, // victim
	"terra17ne9tgk0fatdth272q3n0zkdcrtzagppg9cs7x": true, // victim
	"terra1d8272ztwut28xmjtzrhlu739xpp0g8j4fad2f4": true, // victim
	"terra1ndzjar24ny8aglgc7lfxu9reqs5mjj0vw9cyqg": true, // victim
	"terra1w35qlfw0dhjthzkprwzr62cthedtyps0lqelk3": true, // victim
	"terra1g7dvv3jwf8jr570t795274nc9e9x4rs6m58vfw": true, // victim
	"terra1z2x363rk5v8sh3gp8hes03ket7j065rj2qedfu": true, // victim
	"terra1y5n6ft9pdhc4uj8jz52rxwc2d88aqj9zndeqy2": true, // victim
	"terra1udh25l49ayex99qkllrlsvynmxmslz8vfkmzgy": true, // victim
	"terra1zpx4uarrvxet8twwphhcsz64e3ej8j2awjplwu": true, // victim
	"terra13h0cyduynpnpwy34de8c5ul8zqt7jfdsj527wn": true, // victim
	"terra1q0rsxwpcle05dxxmhx9mxpz4w4q3nxvz0chwl7": true, // victim
	"terra19fvmg86y8tjj8uu2zt2fkdv4amtmlafnk8p246": true, // victim
	"terra1859dahah57cf25h2guanplkll4h8n3zx5ume5w": true, // victim
	"terra1hu3k34run2xshnfjwys48yg9235sagm86erjnv": true, // victim
	"terra1xpgp7d4jh2jnguu7z75fgy4tgytclyu6zhk8z2": true, // victim
	"terra1639vclutk4vtpj6thr308em6pwslu9phgmm7yg": true, // victim
	"terra1rk3tfrr9d2rfq2x5njaugupdsg34nqaefp54ct": true, // victim
	"terra1sg4utsjq06e0ffjzs3q9muuj9yhk6j2h6ftf2f": true, // victim
	"terra1fqrcjukm2y7n9p39khsmkvlxhgmd9jh4aypy7h": true, // victim
	"terra1jhu72gzknscd82kwsw5hv6uxy5au3xvalyq4h0": true, // victim
	"terra1hxmjzfmesvn7zqehf54h2t09vt5w7prgp0xxrh": true, // victim
	"terra1t6v99t29q2hfg2w7j3ejf8fm2eea6wvpzwgmph": true, // victim
	"terra1wa6ndfa47e05r9t225fyx5upnwzlwum3u97d52": true, // victim
	"terra1wxk8t6ah9ncxe3f8emu2l4qnzw4e8c8dfs0wvn": true, // victim
	"terra1uww7g5xrwla4z5q5kfygxc0sq4h5z4tmh3pq6v": true, // victim
	"terra1uxnwyguqa2l6y4m5lljpjrag7yfapj02dt34jv": true, // victim
	"terra1wvf24j8vsreacxfqh66y5cw5h7tqkpzdrt0442": true, // victim
	"terra15g2jaw4dqm40j6hqgegn374xu79awxlypae9fk": true, // victim
	"terra1hc4fgpjrvrq0x5eknzphju5m2y65s0mfscse9n": true, // victim
	"terra1ryhrl2urlyz9kmrux8u5zhcrwdjvgn5wq4ngth": true, // victim
	"terra1kp0drnsxgepvhg5m8ae50e4j46lzd06n895xup": true, // victim
	"terra1zexg70f4vy4e5uue2hstkuadyjhmk89njplv42": true, // victim
	"terra189kmsde6ma56unlmcdzx94dh6u06evpr3e4q8v": true, // victim
	"terra1ydaxacuketkdkaek4crh76wc3n6g6ajedey04m": true, // victim
	"terra1sxc9d6lqsx2f2fay6t8d7pcvhe350v90r8xy2v": true, // victim
	"terra1c496xpwahz580kq4a29vvlsll02zsecpep9pqe": true, // victim
	"terra1p8hrlx9aencsv2k97yurxh2uekq78kr0kztump": true, // victim
	"terra12hhsdw2hgz0sma6drgumuzp8c8sqp6l569h039": true, // victim
	"terra1ku0rlgyt4rja9pr26jshk9hgyxvkj862ka384z": true, // victim
	"terra1jlw9ww2340yqds3z7lda0eyl06qjcd96xlra87": true, // victim
	"terra1nek9h0pudhfzvmupz7n0chuc3892pr69uu0qsy": true, // victim
	"terra1lh3n5petsv23tqaqvfpl6vcezxmepwgt9shy6l": true, // victim
	"terra1ksars4g4tuns78zkhfjjgpptzq2p0ugzuq24v8": true, // victim
	"terra1nh5slclnumjq633cj2l0yk8rgarw74f9wrcjy9": true, // victim
	"terra1fmcvhwv7d0zfmhr0q3ka2gdrldgfud2fdtz5hz": true, // victim
	"terra1fqawxd4wj8qnapagrzktcnrquasegxk8c6czuh": true, // victim
	"terra1pc6ytgynrlglv5unn9t93u73gmt4ghvcrlqq2q": true, // victim
	"terra1pph885lz8dnqqx0urldruu5wu4eq5p6anv8f0n": true, // victim
	"terra1466n83xsavxx38s3qqvvvmd5v5kshs3u4q97jj": true, // victim
	"terra1jle50chkt420pd3p4eykmhwfpx0gvy4s4vxd06": true, // victim
	"terra1umj44wnalve4r5y9xqyj44y0e8rz6l3kjvs4f4": true, // victim
	"terra1847ngg9dhfzlnpl54ur3w2qq9g6cvjg4ft00qz": true, // victim
	"terra1uwvsqvu4mtg4agy4a6lfqqyp4g3mzzhn9qkhnp": true, // victim
	"terra1kaw3p0ruf8gu6mmemvmnrexsjhvv8yqhez0f46": true, // victim
	"terra18szdv3rcmujnnjqams8v5vqlt9spwdcn79z6pd": true, // victim
	"terra1xa0x6jm482vfq8cs9rn6era7c765jvhf8t6rk8": true, // victim
	"terra18yhlwvluchs2qlz56jmvfpykyd3uk06h64ms4g": true, // victim
	"terra1ck3fuacu9p6m5df6gvpygzzzuxn2rmrg7m266k": true, // victim
	"terra1kn5pxutawn9hs8xu62vynwn0ls4rht4c06ty98": true, // victim
	"terra1k3c6v4v8atd3rw8s47yljtwy634czareg2g406": true, // victim
	"terra1jx2dvsxxldl8d2jgc3zpx78ew7kcnkt2mfygh7": true, // victim
	"terra1xucjm00ncvvstz56c24089wkdppnqe56934vvp": true, // victim
	"terra1fhw0cp5sqxgfm4munpne9s3wy8t3aqcpmen5wt": true, // victim
	"terra170e4xmv465phgd7zdt7z5fls72z6cw5zywg08z": true, // victim
	"terra1fezvfgr8qxnsvfc0fwguy2plqlkxf8fjx82m6s": true, // victim
	"terra1g8tk6e5nrjl2ktc5cj8p4q3tfpfpxdrzn40ja2": true, // victim
	"terra1f2fz3wujnxme549pjcwtpuea4cxjnvvr87cs48": true, // victim
	"terra1jgt4pehhpfe4u527u3d284ryste7w6t85lld0v": true, // victim
	"terra17mze6anr5zjk73ex7p4w4am676e6hh4mqlasn9": true, // victim
	"terra1uwqx4sed8xt2cm9mx8hxz5h6s5npx5qd8nwevu": true, // victim
	"terra1pxydsja6zq4s8ksmwwpg5kpspz3lm0ywtl00hq": true, // victim
	"terra1vs5gjqfrgnnkyyqse4prlfv06mvpt2xjedxxt6": true, // victim
	"terra13v8smtad7jx89hv4fjj5j3aq5t9dmvpm3u8l48": true, // victim
	"terra145vf3mahrhcqqdjtz8lmlt4nttl2rlrht6vpt4": true, // victim
	"terra1yyl05auetr8alnwen53e8qdf53js8seyeff9hx": true, // victim
	"terra1nz6p3jastsgplgl4j9smc7hhj9x4ucz5mdzxnk": true, // victim
	"terra1t2qc8hpjyhraark5wd0phxthzt3tpzn3mwksar": true, // victim
	"terra1ua2gkk7jfw3u4kcjsrayytl4xm8nsz0rmveky0": true, // victim
	"terra1hv40yrhsqp7pa8cuqxvks38ujtl3hg785n6yzk": true, // victim
	"terra17ysnpmndwa3xqhd4g8d9fzkan8mjc8hn7aa3r8": true, // victim
	"terra1n62r7q4xjnv252pf90peqr87fptsk9x3a9wp3l": true, // victim
	"terra13he6hrl95nugnapvvw4aeudul842zd6zkdvf4g": true, // victim
	"terra1r0dqhyfhe2mqrahmwx36tlksyk63c5pkr89sva": true, // victim
	"terra1nt2kjm3avdnadusaltjk88ahzj9j7w3nwsusf8": true, // victim
	"terra1lz30qss93c4c6r9zy345r6dep589plrxj5fpdm": true, // victim
	"terra106gvdpvt28jrq27u5uap44ynv8zc44p55uwhe2": true, // victim
	"terra1tus73uwd247nj6qegadzyyzf99nh56w4mfza7l": true, // victim
	"terra1vc4khq3q4mdpjr67rrj045qzqh5nax9vqydfxe": true, // victim
	"terra156vjv2679u3gk6zk02hn7z6svy48r4fy26yrt2": true, // victim
	"terra1uevu5ell48k523pzc0u0wxnx30yuqcmx7suh2k": true, // victim
	"terra192z3zuahqjcepduw3lw7zfpw8g7wfm48yalmzz": true, // victim
	"terra1gdpe8xq8evjfeujfdrq9gxa4x030q9jgp8wysf": true, // victim
	"terra1y38yhzdq82lukupll0d5vz2tsgwtlhk8zykvzx": true, // victim
	"terra18jceq4g0cpa29xwwwfv5kgsqlgmnl0339wwaul": true, // victim
	"terra12p7jt7y0reuwwuuj2rp6kupp87f8l58t7kugtz": true, // victim
	"terra147r6cvh8267q7xhsu4almuhlm3v4v0p7eucuea": true, // victim
	"terra1upsyzmjvv8flatteadymxjm6pgwu6rhp7qpav9": true, // victim
	"terra13k5y7hv6pq4qqju6d5e4v0vqmpcg2nx825m4qt": true, // victim
	"terra12gyxs377jde8l7u6tgd3yde5m0dkd4lralafu4": true, // victim
	"terra1lplr6kyzzlj4r2tdr3zqeyeau90ew5ldemgk8z": true, // victim
	"terra1l4qkmkes2mlrmxmpv4cvckk84r7edry8nmnfuu": true, // victim
	"terra1j48hty4smvr8euc2zvh9ac6crupsehyhwcld9t": true, // victim
	"terra1fhdwpwc0vquxnc749vzn208e9szdtx2h5f2j00": true, // victim
	"terra1vafdle2e7yhvp8xjzxphsu9kx8v2pw67p5ll9e": true, // victim
	"terra1t7nzal75ze3tqp0p6ft67dy5vuv440lyekhfzv": true, // victim
	"terra1wd9mhx2ayt5537wf0n45gjss7pet0kwkdpucf8": true, // victim
	"terra1az87zsectqlfhhxg6c8haumx9uy52eyxzadk3x": true, // victim
	"terra1c6mmng09thtfatrd27yc956qz0q072hh59smc3": true, // victim
	"terra10zqnl8s79qv5yepfltwmcjj0puklqtvw329ztl": true, // victim
	"terra1qdxnactvgunsq6dfhtdrga78kfqfnmyzmmf5tw": true, // victim
	"terra1ndykpj7pfdw4tt7emrxkslm2n7uzm3x48w7mye": true, // victim
	"terra1lud06ncqkcxvqel9sfhl4fe36ggchrnecqs6ke": true, // victim
	"terra19u5293x0k3fytkm0u3prpw900xwznyuxrmvy4d": true, // victim
	"terra1y9l5tlnmwcapalar2xd68e82dtgmtwyv7jznn7": true, // victim
	"terra1p8hl7c0la6lctnyv8canm23f0tllf0cp7wxap3": true, // victim
	"terra14rqplnrly6p0lsjzkup07gskkxt6sa2aerlp90": true, // victim
	"terra1k92w68a3p0gml6tnglryzvm7v46n86jg47g9fy": true, // victim
	"terra1nywja8qykzgcva579tyuarglyv9dc75y8cyw3f": true, // victim
	"terra1jx34m702y0g5ghm6fw7yh38dhkslnvrujqpvdy": true, // victim
	"terra1r0x9vrzn8dlurrv2ra4e8agevf3slsnzgnzel5": true, // victim
	"terra1ztw0f6yv099ve9y68cfha7wc70jvz6c4xpctwt": true, // victim
	"terra1u4zd4g6ts025m6mjy6zqtrcrsy46mkp5almenw": true, // victim
	"terra15eju8a5zhq04zpxnhze28979ktyj582eg4hzjf": true, // victim
	"terra1mv020gvclle57qjadv3wcuurwm2w44tye7znl5": true, // victim
	"terra12rphj3mpuhu6e4a9gpxferaztp787w8fvvkl5h": true, // victim
	"terra1rgtapknej3kqauuzfjjfkxp8xr6nql2drwe7da": true, // victim
	"terra1jzzntf7z5a6jtcamp7edms4ysw5vmrx6lslsk6": true, // victim
	"terra1xffapzxaq57grae9nx5ly38lg2gk2j4epxnwq8": true, // victim
	"terra19drnr8r8uvkw6z9mn38xj40f9g7dyyku8ndc0j": true, // victim
	"terra1d3rwfaxtrkpfy7hj6hqnvc6akcn455nwprzler": true, // victim
	"terra188nwmumfd7puk63yyer6204a0ncs40qjzq4mdw": true, // victim
	"terra16809q0aj5zr49xh0e688cefge64ts95ewzrdc5": true, // victim
	"terra137uwjw8wgu80qpz33r954dgasthtsreezzhjuk": true, // victim
	"terra1vgygh9vytfx4dfus5tltxv6d4xnrkhsy9p90x5": true, // victim
	"terra1jdpcf3r4238z3rtmyfqvqsmryywm7c64cu8eax": true, // victim
	"terra1zpwt0qq8nwfdjs6satufc93ztjfnvck3szha8u": true, // victim
	"terra15zmjvplx6twm8nrvmg3xz6qcqvl8rgncxnjdw5": true, // victim
	"terra13d2ewaj7ajdd5sn5q3x8uh2j9dg23peqjkjet0": true, // victim
	"terra1c42e3sp9ynfm56tdrccqku9zgymzh4kjgzh229": true, // victim
	"terra1vxawyk2ewxysxldu0x3p8xw7974p97uvftqccv": true, // victim
	"terra1qreryuj7ernp7cl38504vyl20de3uj2tfjg84e": true, // victim
	"terra1p52uds89u2r65wg66s9x27y0c25suk4xydgpp0": true, // victim
	"terra1243u940c56rpl9q85enfsxdk8uxfl72805uzh5": true, // victim
	"terra1p9hujqrtsfmcme8x84qptnurcc3fmlvs5ez9kn": true, // victim
	"terra1adcfwdu9429d97y7k4fj4lygtr0g5l75ckrsa7": true, // victim
	"terra1hm7u5rpafy64lpz2gya4uykdqxjrlpl0svcjgl": true, // victim
	"terra1cuu4q5fdpaxuwxqxkl7vzvld0msj6k0q2r9urw": true, // victim
	"terra15xmk9gnt867p39y2np9w34uxa58gal4h97sewc": true, // victim
	"terra1cws9l6jk3aslh5rqjkzcc7532qwg93ec8jg9j6": true, // victim
	"terra1rc4ltl3a2lamwrmnd9xge4unstk2p8fvs7kpjd": true, // victim
	"terra1lrnfvkne2l2nxa20963u6ag26m5cvzhs9wfhru": true, // victim
	"terra1026qm40cwyg32qhhsx923lkxvzm7eynjwv8mza": true, // victim
	"terra103ra79dl2un2ltknhyz7crm5y29g4vhmyh9nuk": true, // victim
	"terra1rdvx0c9hnfeyxc3cwt4nhtylcnm4ttx6wu6wrp": true, // victim
	"terra1wk29zk9alhg8qjngqtj83g68xhsxrx89pdtke4": true, // victim
	"terra14he8dj2rw9eu9v7kkmrzewrnduu9tpmrpn0zjj": true, // victim
	"terra1ru8ckf5e3ufpupu9l7t9fg44vqxg69fgt6dj6d": true, // victim
	"terra1pu6e63d3rd7k3j29zd9v2d3r37z7mulawq4k88": true, // victim
	"terra1ge7cl3ta85l0g9xz0thknmlenu0wtd5qr9nrfs": true, // victim
	"terra1esfyfxqtpeyhgqn7z780hslnavjsj65sktq4xe": true, // victim
	"terra1ndp0e0udcl8p05st25t03q976uagp7wsn00dly": true, // victim
	"terra1kkez97emjcndk2j7jt4srruzgrp30x0tjhvf8r": true, // victim
	"terra1grvypdd2p79swxyznenngenmsapfhp39rxjpuj": true, // victim
	"terra1drg0w8len8vwrd30luwsp6rld8wcpk8cjsshm8": true, // victim
	"terra1wj0aqr2x9qe8leje5wwkhv9p2fg7hzjwu9uj0a": true, // victim
	"terra103usrxxz8tedl3lphhf250s22xr836jxap7kfa": true, // victim
	"terra154mux25dply66hyeag4dzra636hf55q8m2v8w9": true, // victim
	"terra1h5rc3nhk99t447m6546eg7mlrva7s6ye5zs5jc": true, // victim
	"terra1a89hu63z478ra9eefyntunkg7sed5t437drr8r": true, // victim
	"terra1sujsalpvuq09vhey52xgxmnw3axarwew8zgr86": true, // victim
	"terra14ums9ud62tqf9l6ytslu55w6f78acvzzn39qny": true, // victim
	"terra1r0shju4emnmlyseymlntqauz54w9ckfnhu6ja7": true, // victim
	"terra1jc79v52k4q5ypuwdyk020clj3fpcvahn97s4yy": true, // victim
	"terra1q8yyzed5432zw8qqtfl0zc6qrgc84q8n429u0y": true, // victim
	"terra1tswwm5hu50kmjx4ul7fz3aglec446xxv20ta7j": true, // victim
	"terra1pe9gxszyrmrqjfj5gngds672de93h6zf9pyf9t": true, // victim
	"terra1est7pkfyycwvxketz2c9nh6gqkf6vu74ljtqwe": true, // victim
	"terra1j4x0r9rl9ppkkhn2u5tfvdmp4ms8wup7gs04ld": true, // victim
	"terra1f3c8z93eyznr4pg4sq22kt2z2k6dkxhfw3edtu": true, // victim
	"terra169hpn0fr2tf5un4rknndttncveshqkkq3yp6x8": true, // victim
	"terra1rvgdarl4gz33qtkhm3r0u0vh2xsfsdsfm5yprd": true, // victim
	"terra1ddpuzl4x00mmd2pq9f9pwl0dp9qc8rphakypff": true, // victim
	"terra1h8atyz9derq3enue2gqhgxlptd2krm0vm9kwvs": true, // victim
	"terra1fh6e62vagqpjkhg465yx8h4226chklsc5kqrt0": true, // victim
	"terra1ecglqs8zn4cuglqp78dc0suynzanue0ga3nj4n": true, // victim
	"terra18w2mxw4mq8dv86dcpesghflus54aul6mzrwutl": true, // victim
	"terra177k47d0czu2h8wg307amlken7q672ccd3h3hyq": true, // victim
	"terra15h3alfgw8d0wq6ysufrwthrlhgxefxnly9r70m": true, // victim
	"terra14h853rvlsemv9pkuhjke7yry6mry4yh0rkuhcd": true, // victim
	"terra1vmdkcmt0mg7x0lhv3ls7v29km4w66upf04wpkr": true, // victim
	"terra1m3tnq5rlhrhytzg50vtj0mykf9vnfk3hqj7sr2": true, // victim
	"terra1dcaxl9zqxzrfzqexler2jsv2mphdq60k77evxe": true, // victim
	"terra12fqcax6ukq9cfnqml3yenxka8ahmwy7zhe66e0": true, // victim
	"terra1heqm272p9u2nkrea3cttps3xe0t2lqwul8gvzz": true, // victim
	"terra1t6wphl9snmflj30ernt59uxl68nms6jfua5v0s": true, // victim
	"terra1r502l89gmk2tdwz3xmk8d4kgutgcnz3nely7qe": true, // victim
	"terra1lm536xtz6ynenatpw3x8t6ane7xqd2qdmdaqw4": true, // victim
	"terra1sty6xykusmtm08k46e9usm9sssye6zzpztdalg": true, // victim
	"terra1htvl5c3p7g4rmatl0r30n9zu285jsffh4pcss7": true, // victim
	"terra1z6y8lz9pn93m6ekn89v2t2x0pm9expngv6sq5l": true, // victim
	"terra17j8v8a5hgp07kkyxrdjklkx02c7uhzphmxz6xv": true, // victim
	"terra1pwdvj573c9vrfkuh9ueypdkpuzejn9qt3vjcju": true, // victim
	"terra167wh8lwl63qhl2rsqpkya3xyvvwcesy3lavrwg": true, // victim
	"terra12h6ukm2e5t679cndyw249gvapnhkmz4k4jjmwz": true, // victim
	"terra1w5kkc4he8xzqewdc0xmj8gq0wwsush2y89wxm5": true, // victim
	"terra1xuuxtm38aj54juz8wln35fpzy4j6xu8n9d76rq": true, // victim
	"terra14wzpp09huthk3jc8hqpknkm0jqjrxfm97ag2f9": true, // victim
	"terra16gurc6zr3gvpngusnkh0urs64fh9hc0q34pynr": true, // victim
	"terra1zqv7nkpdkn09dt89xj3zxvu3frd9r0eummrg92": true, // victim
	"terra1l3j06ve5k6g0sazhsu96ngl76d2msdagfrlg77": true, // victim
	"terra1ru034204p8mca3kwzx8kdcn4vmdrxkrw8mpcl3": true, // victim
	"terra17egpwu94m7763rf7vg2y3evyfygwl82qtqaehc": true, // victim
	"terra1zqj4ecjy3d8hy9xs0c7kws76vkfh3nf00anx0v": true, // victim
	"terra1atnqawfeg09ym0nkldjlqmcvssnsdm4r4r2jzk": true, // victim
	"terra16pmh8d6jz82pzjn4ssfeap3uww97h68d0g89tk": true, // victim
	"terra16j88z3dl48lktvzddkr0uzmcgtyqxh76hwdxwv": true, // victim
	"terra1hvdspxrhdmazv8fqm03049artr48e4mlz9aapw": true, // victim
	"terra1ct3zap8l7z7afj4tewwk8pengtqupsge5hsnem": true, // victim
	"terra1a6jjy63ycq75p0nnf3xrrnecmlxy6ltm3qncr3": true, // victim
	"terra1kdef0fwfyqwe6pl22aze0yvkkprpl7q7a256m2": true, // victim
	"terra14m8unq627j97ysf8k0vlm6nukuwx5wd0zt7q5m": true, // victim
	"terra1z3skwf6sk8upsj25rannsmyevreg2tg0w6uey6": true, // victim
	"terra1qjdckuj44jh60zdlam9d4ld0wwkuv0rk742ect": true, // victim
	"terra1fs03yqxa96gmca0ut0pc7l694n94p086h0x34r": true, // victim
	"terra1t5cster2amkt644g9wc2yck8gqk9394qgptc68": true, // victim
	"terra17yxt53q64u85cepkcxa8tf84csqrl99j0uge6v": true, // victim
	"terra1fl4g4hugsrsgca75jmym4hdp5nf2edvpe7wdhc": true, // victim
	"terra1jumfgsyf788uclzju0xk9tmvdh60k993ljhm8p": true, // victim
	"terra1pwp6ngdkgws2qemx423kn0mxgkgng56pxvh2am": true, // victim
	"terra129svdvqlntq2mcklwzkncjzdm4lnstqs5lpqma": true, // victim
	"terra18ag8tjt4kj6qzzhsagq2n3nqs2amlykuwz90mq": true, // victim
	"terra1ttgwj8emn08adtsj37wudnxne5jkqshx302h3p": true, // victim
	"terra1yujmhgv9lutp5ffekf0mnv9afyqpyuufysv7cs": true, // victim
	"terra1hul762jhf0fzxmzwwjnt5ye398h90fwjav2602": true, // victim
	"terra1wqsljzwltnng64xgajh7dpadhngajh6zkrg8qd": true, // victim
	"terra17v7v4g2cdjmqxj9jc284zphse3kw43rlwn4d9m": true, // victim
	"terra1akpz8efgy55c8amt6fjkh90m2zenmzmwgywcux": true, // victim
	"terra10ts9muxs8epka9uh0fraldg8m7a3h08vpg6uvs": true, // victim
	"terra1q74dhv0a9adppvw85sg8whjfc8nx9sk0hknayt": true, // victim
	"terra1k739mdg25r0chpufhu7ga56lj9zu59jhqad0d6": true, // victim
	"terra1mvqz4tlcm24zrmccxpekgh8kxwpff85msv6nmj": true, // victim
	"terra158emqhvlv6ft9qh6j2sqk9whssx0ept9j8de7w": true, // victim
	"terra1cd56jzvg58dgw5zdj5k8upjs7fwk6q3wwh2sy9": true, // victim
	"terra1s7wkcrjjt33sm45u5x5uxs2z7pdugv0kes47ya": true, // victim
	"terra13prcf2q936takv75x5g7fnkczq7m2s5efgsswp": true, // victim
	"terra1k39fz62jerndn7d2m22y48up27h0jg9krjeqts": true, // victim
	"terra1at435vrq82njp8xdnktg6emnrtysefe0dts0aa": true, // victim
	"terra1fmlqg5m38juf9nfe5amuwjlqv8h84xqhsajk3s": true, // victim
	"terra1fg85pg27p6g8wz9s6nn5crayqklmnj56q9l9dj": true, // victim
	"terra1vg4tp9vetnjg9shsxrcqmjrtswwf4ewtu40rcu": true, // victim
	"terra15efhuneetf4w7j92rnftgetq98am2dxkeh8vhh": true, // victim
	"terra1zdh7ucs89leyv2vxvyqpwxgmr4jayw8z7hsre6": true, // victim
	"terra1rmt59lnh55dysl3vrqw3j8z3l2nm2lj7kc975r": true, // victim
	"terra1st4v5tyg7e8phqtzgue2ljfcpk5axu0sn56dfv": true, // victim
	"terra1nla6my64emjvdejvd060wa9rjlw4j637ymu0c9": true, // victim
	"terra1v0y4qafhxzjv7le6stnzfhflu3agvdllmcnwnr": true, // victim
	"terra19g6dkdkjjavl8j39ud8fwm7wunqxk7yju6elna": true, // victim
	"terra1685qe3cjwjzn5vmysvjcgupqhfdv5xywpcc3r9": true, // victim
	"terra1u5gwgzres6jmvmg4q69nkd93zmwxxtcfynvmf9": true, // victim
	"terra1ls85fkmjzdqj7avmpmc27xpvca806jp5sdepx2": true, // victim
	"terra17j3ygjr43slvwu8sq6zn9x69psmfzh7aykzpe9": true, // victim
	"terra1gjnrxkd6tc2d9dr0a4fryzq99hwds570pmnw7a": true, // victim
	"terra1q3996m2fx8r9ywnke3fk5rzawzl5dd4g4jgapq": true, // victim
	"terra1qhxpsds2tlr2wrq864w40qutd9f9puz9gukrq5": true, // victim
	"terra1mlnrqj0p6agnc638496d26wgtqe4yfrzsg6rfm": true, // victim
	"terra17cxvggu06f6p2flv9rnkmmgx3nr7teyjrfqysd": true, // victim
	"terra1c9hs668ee4rk5nacx2tygd3tgv0a9ph28zt4w0": true, // victim
	"terra1ymkvtv4u29a3egazpyn3c2049q20t6qzyt66xa": true, // victim
	"terra175uue3ztzdemameswa7dpu60asrauzc8u08kmn": true, // victim
	"terra1fywalygs37vcsrcm4c224n4ltt8mhwzyx4ts47": true, // victim
	"terra14kju49xmac6jm5gllncuf0jc36ds08afwz4r2d": true, // victim
	"terra1wrre425dtpwhr6j9rt8hnqjjs2l30zm82h03vg": true, // victim
	"terra19yhkvrrxrkmkktnux827phkc43rmpqfyy7nqd6": true, // victim
	"terra1flz89pkz5xsrlnxn3vjayhplvlfygnummdkr3e": true, // victim
	"terra1ldff05jhweptzyae32gmaxse2uufkl3l3qv028": true, // victim
	"terra1c5rzn6q9u5sjqju8d3m8jxmd0nkdj39eg8h2en": true, // victim
	"terra156h3w29vp82j8z67hw64zamljjvk8qgcjsp0l3": true, // victim
	"terra1uhlfvq7sq30mfgygjz5nrtffdjgcxmmn3g2xyh": true, // victim
	"terra1kgp3c7cf0pr8esypfe5jx8p5yf7eju0hx8wrm8": true, // victim
	"terra1v4cmzdx0vpjhd4snher9zg86k2pcsn3lymqfcx": true, // victim
	"terra12fg45a60gyv5hnt2wzry3m0at4enk4y9rhsauh": true, // victim
	"terra1t09f386rauta4vdq7ujqfkgvnevycfgcnjgjeq": true, // victim
	"terra19quq95fwakn0s8zegc5ah0q4ut4yzplq3kzjy4": true, // victim
	"terra1pjxtkrzp7na6z2ddfgmgf65z8e33cc69wzl39m": true, // victim
	"terra1u22j98d5095dkpuu432ew460smrzjjkha9mry8": true, // victim
	"terra1mff8zqxruphjsp96th9k7eehvn6z6j00lm8gpe": true, // victim
	"terra1sa7c3jqlee3e6phh4ld2pk79kd87as5thdkl2l": true, // victim
	"terra1u4emgdrcr0mdyykde6jx62fpzhaaugjrhxfn6p": true, // victim
	"terra10d4zw22qzpwrtnj0fh8tpqgs6jxnkenzj4k7kc": true, // victim
	"terra1j4cmq425xmg6ryl7fhc5jr2xjeyd536fdx97jj": true, // victim
	"terra10nu2a4qvy045jj4gy2tn2z49pnt3vw0qw8tats": true, // victim
	"terra12qeqrw3jk5z45kzz8gr5kp6jgj02szxewy2ps8": true, // victim
	"terra1v4h6vprm59ktvtj0xu7wvz53jg6ar6rdcsx9pf": true, // victim
	"terra1wvkkrpxk73kqnz882fettz3nq8lwymx50wq87w": true, // victim
	"terra19v2wzwxeay8acl0g5rh0q48c70r9escf62a7uz": true, // victim
	"terra1dqvh6aynx6q0mxuatd3apderrx9uq4p9qf8rjt": true, // victim
	"terra1nvsz3lhj68nmmcp3t23gc8c2fjpeeyy7lqyt3j": true, // victim
	"terra1xaxrgfk5j5z7tqtpngwyfjpeeq8t3el37y29ef": true, // victim
	"terra1veez3f3mgm27se3dj55k037ykvzmdy08smaztp": true, // victim
	"terra1nat3ewjhrht3qu4hmzymazdfzw34zx5jh674xq": true, // victim
	"terra1myr6237pnzghcpuk9zmxtvld3ndkrnyk0y2ex8": true, // victim
	"terra19nekp9vnnjq28n4t4lydmfctunmhdnvaz0hv86": true, // victim
	"terra1qme2ehhdkc3j665dcq4l2xp4qcztl3wg6lej4y": true, // victim
	"terra1yazda3hv6hk7y53m4nj5e9eqpnrmavcrktcs73": true, // victim
	"terra1qr3dx4qh7xsewkkg80e6ehq295fv6upsk7w7x4": true, // victim
	"terra16se22xlfp3zz53ateqqtja3ut2h9r9z9vwf0nk": true, // victim
	"terra1aewek4geun6hj9zs6tasd3khzjld58vgkn6dhe": true, // victim
	"terra1q0rn233yyqttx8ftnwal4v9yljwd4008a3fy08": true, // victim
	"terra1mrxlx2pxya49lhklgy9kr5lwx0x2t7c9j0rgh7": true, // victim
	"terra1raupnyvhc7nj0ar02w3q8gp7q0nnjakh608syv": true, // victim
	"terra15xp3e8zq5f2wdswhev40u0k2k0wf9mv6g50fs2": true, // victim
	"terra10qu3zracgv4qvewejrj0xr3nfuc88x8k2nzwx8": true, // victim
	"terra1dp8lncleunqqxj62xz95nd0lvjzjf7kzkagzv5": true, // victim
	"terra1gvrggs8rukmg37p35mp773mrvfeklhwwa57x0j": true, // victim
	"terra1a49p6c4pn3t5ut87gtwg6vw7pvwgjszagte8g8": true, // victim
	"terra1f5hyt3tjjwe0xvrlnd0d9znrd9wqyezcjzl39u": true, // victim
	"terra1gg6tp85tdku6nq347ak3gzcpm7w6ehteyf6z36": true, // victim
	"terra1wuy7ywvwxccmh6xc4y2j75m64wfgq4ds07qz22": true, // victim
	"terra1r90qju3602j3y5qy7q966xjnuzff4mhs6zl2xh": true, // victim
	"terra140q9vp3e2gzmkxwjam2vfdl03rpjcrm57zv5ta": true, // victim
	"terra1uwjf55qayjrkt0pgw8w06rsrlqy7jnpfu46uz2": true, // victim
	"terra1snqazxpnrz7vq77k43hjz5glqp3ujuy7my925l": true, // victim
	"terra1grgnnu8lxp6m89dq2wp454hj0hgahm0dl085y4": true, // victim
	"terra1wl3zhj49v2l94ggykmhvcttdusv3gzu7m6pcq5": true, // victim
	"terra14grum25hp89ueanzupprqwpu76vy647p6c642d": true, // victim
	"terra1xmhplxe322q4vlwe9cqmz8vf9utarlynu5ufta": true, // victim
	"terra1vw76jgq4yj2e6xaqswpf770wh6euhl6n6yjggy": true, // victim
	"terra1jws0fx6zhgcpe8hgvw4mr7vy9q7ha4avv3q0mk": true, // victim
	"terra1chs7sjjjwuudrnetdeshvsvz636a6x2au3durl": true, // victim
	"terra1rwquzhyxxsl6wc0put4gnk8zt57s9gzche4gzm": true, // victim
	"terra1fwmxqep3yjq9mqp7xwexcrhth9t2rtvmum9s7k": true, // victim
	"terra15jq22ust7lwpqpw74a2nm5wz3mzjystssjpeu3": true, // victim
	"terra1lxrf9d3crczqg6cxlafgk4zqznjnhq9cyf7sem": true, // victim
	"terra1pkn5cvvm2n2srqgf274k3xdkxmpyjlj5sdr6zk": true, // victim
	"terra1489ryykf0ju62q2xr500u5rmwgekwgm5l4v8mv": true, // victim
	"terra14um45x52z3mzxzgtjqfqhw9wx45fj5hnwre8ze": true, // victim
	"terra126pwqwszfyhpmdcuzt5evmpgc0ynn6wevrk9el": true, // victim
	"terra16jpfjxvavajaml6y8hzcz4tlqvgrch0zdknm3e": true, // victim
	"terra1lv3dkedzw5tnn2982r59lh6fn0xlua7wyn9vrn": true, // victim
	"terra1c2aqpxs9k2h53z34vgzn9wwdeuz6ldj467fs58": true, // victim
	"terra1u8czdttfsevzf2gnze2vw9klgdr6hsrp2nj8t7": true, // victim
	"terra19gz7fm55v50rzvre2e3xx08ee4add0lu6hua8m": true, // victim
	"terra1ph7y7cp94mtpdhdjuwn0gckcxa5v02rcpjapxl": true, // victim
	"terra1t2d3prfx5gur57hp5zt0c9dafjw0vxd0axs537": true, // victim
	"terra13prtycjnmmqayfpwl3z2qqlrcew8jztxcy0fhg": true, // victim
	"terra1dguajk3gt53dhezrw28lw3rjq7qpu0a9c4h6sq": true, // victim
	"terra1agslnwulgdmge80ljrv3narj4u3ssdu8uwrsh8": true, // victim
	"terra1h40v9mwqggdpj6vsmgtju23ux0zwks9q78w2y9": true, // victim
	"terra127yfdgj8ursvj9p7ljc9rfr8ty4y90d73hpleg": true, // victim
	"terra1c9w5snjunchrt73q9sq3gkgjdvfykjj9flapyj": true, // victim
	"terra16r3hh5dt2w2hgwwwxkwwpkqeyzaqtqwjrfwldn": true, // victim
	"terra12p75pzwnwpy03ydkezu2rkc6q8whz2n2607x0x": true, // victim
	"terra1wyhlxg3qk2dxw776a4pvyr77yslxj2p7lx4n97": true, // victim
	"terra1pvs4rrn6pr4nn27h8vprde0dwld3mrre7ucywr": true, // victim
	"terra1e5ncelsh4qhqt3s97vn43hxlhmt7zd43yszdnf": true, // victim
	"terra1sl2tca3cp930nnqqqsyht3gn2jacj80c62l8kz": true, // victim
	"terra135cfxn5f4hyn6yd40gtq753qxje8c0y8f0n7ku": true, // victim
	"terra120ysvgl83jqmu7h2qnzq2ns4csgua44fhj42ar": true, // victim
	"terra1l7u75twcf7negcxvflz7908m0alq6ygl4p2xjd": true, // victim
	"terra1zsky63r58vc7dfn3ljj32ch6fyn4e5qd8skzyz": true, // victim
	"terra1u870nc5kwt5ntsqmjd7jvuv79c7ua4awx6pdsz": true, // victim
	"terra18jfdzcfvp9jp08wadu42cqw843ds9zcr7fscjr": true, // victim
	"terra1yrrlejd8qec6gpc44raulwy2uayq5nzy3u6jtz": true, // victim
	"terra10alauxa78e46mtr75r92w9eznvm8kyxczv46v8": true, // victim
	"terra1wxzfajhhm6l79y32ud7nv2up837529em8cz75h": true, // victim
	"terra10wxsngdf4v7k0achcktjwvt6tgwpeqlrnqav5q": true, // victim
	"terra1y3sl04cxk5u0z643lep35wtpzfdddu58mnz79m": true, // victim
	"terra1aprl45fx73m4tws87ydql5570tmu0u3zrynq8a": true, // victim
	"terra16u3gjst0h6cn4e4hnkg3lzyw32aw3s5g7j6wmq": true, // victim
	"terra1qmrzf2hzytf6fpuppl4qjs79mev92c5ecch2j2": true, // victim
	"terra1xzawkluxh7lahk9havmj38eacukvrran3dhakf": true, // victim
	"terra1rf58rw0cnzz3f74g94l5jy0mcv9nsekjk40lkh": true, // victim
	"terra1agp4wwzgn6fuxqrsqgjvfhvmfqf63ekvn9nvyh": true, // attacker
	"terra1dmfp9wz7uguz99sac777wm6pdmkm7d2vjg5vda": true, // attacker
	"terra12jkvks5d4q8x4pc02aqvd8ljfhqc3da9y3ern7": true, // attacker
	"terra185wxs6jhxe4lljqv956e777ds364muyjdsdn22": true, // attacker
	"terra15ujjdtez8vjn0r5emt92wur5edty82z5epp366": true, // attacker
	"terra18x3zq0fnkxazte2dhwjapagd5kh42f7dmh9je6": true, // attacker
	"terra19jkevz5u2fzqszsat004t8va65x4sjxe0xjylh": true, // attacker
	"terra1kugad5384pmulgpwu3596n2vdcsnk2mh3glxjr": true, // attacker
	"terra1f5ycdjunqcunzhfklww0duuwkeve92czktn5za": true, // attacker
	"terra1vgrkmad4hhyfnyw8v8d3dh5u7glvdt83a0lmu9": true, // attacker
	"terra1xj3t98tkrf5ynwrlzp4ss02snmwen4523z9659": true, // attacker
	"terra1y0hxjauxge52ee5uq2lnx9kduc3458pd98ttdt": true, // attacker
	"terra1he49aq2dq6wqyts0750x7jalyer99pka9aq0g5": true, // attacker
	"terra10e38wnyca7h8u8gq3rmt73kj9q03dhgk8yudfv": true, // attacker
	"terra1dqwpehuzv0759evv2wc2wyn6vatenjxjh55hk3": true, // attacker
	"terra14twqn6estwc8zqpm3t37glwc76t39dytg5gq4q": true, // attacker
	"terra1pgsp7l5y983hx3t8jmjwq34dptjnghk4jessz9": true, // attacker
	"terra1zsdyv4azfdnuqkrlp3npa705sjdttm9aj5qrvf": true, // attacker
	"terra14hwe4te0aeyg7wtatxc6xk2yd420e7zyu076pw": true, // attacker
	"terra1xa3pmdxxudegk9wysvjundsknlaa6frgs864vz": true, // attacker
	"terra1uaspcnh3r5etr3szfmgn6ecuum00dm3l7fstvz": true, // attacker
}

// This account is used by the v2.22 chain upgrade simulation only.
var testChainBlacklist = map[string]bool{
	"terra1a698u5rm2x6y50x5m3q37tnn0k6d4rjpfc8e7h": true,
}

type BlacklistAnteHandler struct {
	codec codec.Codec
}

func NewBlacklistDecorator(cdc codec.Codec) BlacklistAnteHandler {
	return BlacklistAnteHandler{codec: cdc}
}

func (b BlacklistAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if ok {
		signers, err := sigTx.GetSigners()
		if err != nil {
			return ctx, err
		}
		if err := checkSigners(ctx, signers); err != nil {
			return ctx, err
		}
	}
	if feeTx, ok := tx.(sdk.FeeTx); ok && len(feeTx.FeeGranter()) > 0 {
		if err := checkAddress(ctx, sdk.AccAddress(feeTx.FeeGranter()), "fee granter"); err != nil {
			return ctx, err
		}
	}

	for _, msg := range tx.GetMsgs() {
		if err := b.checkAuthzMessages(ctx, msg); err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

func checkSigners(ctx sdk.Context, signers [][]byte) error {
	for _, signer := range signers {
		if err := checkAddress(ctx, sdk.AccAddress(signer), "signer"); err != nil {
			return err
		}
	}
	return nil
}

func checkAddress(ctx sdk.Context, address sdk.AccAddress, role string) error {
	encoded := address.String()
	if Blacklist[encoded] || (ctx.ChainID() == "blacklist-v222-test-1" && testChainBlacklist[encoded]) {
		return fmt.Errorf("%s %s is blacklisted", role, encoded)
	}
	return nil
}

// Authz executes messages on behalf of their signers, who are not tx signers.
func (b BlacklistAnteHandler) checkAuthzMessages(ctx sdk.Context, msg sdk.Msg) error {
	pending := []sdk.Msg{msg}
	for len(pending) > 0 {
		last := len(pending) - 1
		current := pending[last]
		pending = pending[:last]
		exec, ok := current.(*authz.MsgExec)
		if !ok {
			continue
		}
		msgs, err := exec.GetMessages()
		if err != nil {
			return err
		}
		for _, inner := range msgs {
			signers, _, err := b.codec.GetMsgV1Signers(inner)
			if err != nil {
				return err
			}
			if err := checkSigners(ctx, signers); err != nil {
				return err
			}
			if _, ok := inner.(*authz.MsgExec); ok {
				pending = append(pending, inner)
			}
		}
	}
	return nil
}
