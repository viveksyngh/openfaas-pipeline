"use strict"

var Minio = require('minio');
var sharp = require('sharp');
var fs = require("fs");

var endpoint = process.env.s3_url || "minio";
var port = process.env.minio_port || 9000;
var secretPath = process.env.secret_path || "/var/openfaas/secrets/";
var accessKey = fs.readFileSync(secretPath + "s3-access-key").toString();
var secretKey = fs.readFileSync(secretPath + "s3-secret-key").toString();
var destinationBucket = process.env.destination_bucket || "images-thumbnail";

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
    var parsedContext = JSON.parse(context);
    var bucket = parsedContext.bucket;
    var objectKey = parsedContext.objectKey;
    console.error(bucket, objectKey)

    mc.getObject(bucket,
        objectKey,
        function (err, data) {
            if (err) {
                console.error(err);
                callback(undefined, { status: "failed", message: err.toString() });
            }
            var reiszedFileName = objectKey.split('.')[0] + "-40x40.jpg"

            mc.putObject(destinationBucket,
                reiszedFileName,
                data.pipe(transformer),
                imageType,
                function (err, etag) {
                    if (err) {
                        console.error(err);
                        callback(undefined, { status: "fail", message: err.toString() });
                    }

                    callback(undefined, { status: "success", message: "Successfully resized and uploaded" });
                });
        });
}
