const express = require("express");
const router = express.Router();

//########################   BlockExplorer   #############
const Blockchain = require("../traceapi/blockexplorerapi/blockexplorer");

const DriverApi  = require("../traceapi/driverapi/driver");

const MaterialApi = require("../traceapi/materialapi/material");

const ProductApi = require("../traceapi/productapi/product");

const RetailerApi = require("../traceapi/retailerapi/retailer");

const SparePartsApi = require("../traceapi/sparePartsapi/spareParts");
//###########################    use  route     #########################

router.use("/sparePartsapi",SparePartsApi);
router.use("/blockexplorerapi",Blockchain);
router.use("/driverapi",DriverApi);
router.use("/materialapi",MaterialApi);
router.use("/productapi",ProductApi);
router.use("/retailerapi",RetailerApi);

//###########################  exports     ###############################
module.exports = router;
