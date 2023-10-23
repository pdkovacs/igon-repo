import http from "k6/http";
import { group, check, fail } from "k6";

const sampleIconfilePath = "/home/pkovacs/github/pdkovacs/iconrepo/test/load-tests/demo-data/svg/48px/attach_money.svg";
const format = "svg";
const svgFile = open(sampleIconfilePath);

export let options = {
    batchPerHost: 6
};

export default function() {

    group("Create and load image", () => {

        let newIconDesc;

        group("Create icon", () => {
            const iconName = `attach_money-${new Date().getTime()}-${Math.floor(Math.random() * 1024 * 1024)}`;
            const size = `${Math.floor(Math.random() * 10) + 1}x`;
            var data = {
                iconName,
                format,
                size,
                iconfile: http.file(svgFile, `${iconName}-${size}.${format}`)
            };

            const resCreate = http.post(`${__ENV.ICONREPO_BASE_URL}/api/icon`, data, {
							cookies: { mysession: __ENV.MY_SESSION }
						});
            const checkOutput = check(resCreate, {
               "is status 201": r => r.status === 201
            });

						try {
							newIconDesc = JSON.parse(resCreate.body);
						} catch (error) {
							console.log(">>>>>>>>>>>>>>>>> ", resCreate);
							fail(`unexpected response: status is ${resCreate.status}`);
						}
            check(resCreate, {
               "icon desc OK": () => newIconDesc && newIconDesc.format === "svg"
            });

						if (!checkOutput) {
							fail(`unexpected response: status is ${resCreate.status}`);
						}
					});

        group("Load image", () => {
            const name = newIconDesc.iconName;
            const format = newIconDesc.format;
            const size = newIconDesc.size;
            const resp = http.get(`${__ENV.ICONREPO_BASE_URL}/api/icon/${name}/format/${format}/size/${size}`, {
							cookies: { mysession: __ENV.MY_SESSION }
						});
            check(resp, {
                "is status 200": r => r.status === 200,
                "file length is 443 bytes": r => r.body.length === 443
            });
        });
    })

};
