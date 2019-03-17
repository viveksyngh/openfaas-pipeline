"use strict"

var Minio = require('minio');
var sharp = require('sharp');

var endpoint = process.env.s3_url || "minio";
var port = process.env.port || 9000;
var secretPath = process.env.secret_path || "/var/openfaas/secrets/";
var accessKey = fs.readFileSync(secretPath + "s3-access-key");
var secretKey = fs.readFileSync(secretPath + "s3-secret-key");

var mcConfig = {
    endPoint: endpoint,
    port: port,
    useSSL: false,
    accessKey: accessKey,
    secretKey: secretKey,
};

var mc = new Minio.Client(mcConfig);
var transformer = sharp().resize(40, 40)
const imageType = 'image/jpg';

module.exports = (context, callback) => {
    parseContext = JSON.parse(context)
    var bucket = parseContext.bucket;
    var objectKey = parseContext.objectKey;

    console.log(bucket, objectKey)

    callback(undefined, {status: "done"});
}
