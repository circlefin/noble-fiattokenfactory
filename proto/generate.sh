cd proto
buf generate
cd ..

cp -r github.com/circlefin/noble-fiattokenfactory/* ./
rm -rf github.com
