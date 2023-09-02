node_package_name=node-v8.11.1-linux-x64
node_package_roll="$node_package_name.tar.xz"
expected_sha256sum=6617e245fa0f7fbe0e373e71d543fea878315324ab31dc64b4eba10e42d04c11

curl https://nodejs.org/dist/v8.11.1/$node_package_roll > $node_package_roll

sha256sum $node_package_roll | awk -v expected_sha256sum=$expected_sha256sum '{
	if ($1 == expected_sha256sum) {
        print "SHA256SUM OK"
    	exit 0;
    } else {
    	print "SHA256SUM was expected to be " expected_sha256sum ", but was " $1;
    	exit 1;
    }
}'

tar xf $node_package_roll
export PATH=$node_package_name/bin:$PATH
node --version

cd client && npm ci && \
cd ../backend && npm ci && \
npm run dist:all
