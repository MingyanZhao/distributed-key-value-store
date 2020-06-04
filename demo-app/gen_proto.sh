protoc -I=../protos ../protos/follower/follower.proto --js_out=import_style=commonjs:src/ --grpc-web_out=import_style=commonjs,mode=grpcwebtext:src/
protoc -I=../protos ../protos/config/config.proto --js_out=import_style=commonjs:src/ --grpc-web_out=import_style=commonjs,mode=grpcwebtext:src/

for pb in $(ls -A1 src/follower/* src/config/*)
do
  echo -e "/* eslint-disable */\n$(cat $pb)" > $pb 
done