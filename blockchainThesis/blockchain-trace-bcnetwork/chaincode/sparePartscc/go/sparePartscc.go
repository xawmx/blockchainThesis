package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type Crops struct {
	//物料ID
	CropsId string `json:"crops_id"`
	//物料名称
	CropsName string `json:"crops_name"`
	//物料类型
	Address string `json:"address"`
	//合同编号（国网经法）
	RegisterTime string `json:"register_time"`
	//品类编码
	Year string `json:"year"`
	//排产计划编码
	FarmerName string `json:"farmer_name"`
	//种类编码
	FarmerID string `json:"farmer_id"`
	//原材料厂商电话
	FarmerTel string `json:"farmer_tel"`
	//名称
	FertilizerName string `json:"fertilizer_name"`
	//状态
	PlatMode string `json:"plant_mode"`
	//计量单位
	BaggingStatus string `json:"bagging_status"`
	//采购订单周期
	GrowSeedlingsCycle string `json:"grow_seedlings_cycle"`
	//销售订单周期
	IrrigationCycle string `json:"irrigation_cycle"`
	//排产计划周期
	ApplyFertilizerCycle string `json:"apply_fertilizer_cycle"`
	//生产订单周期
	WeedCycle string `json:"weed_cycle"`
	//备注
	Remarks string `json:"remarks"`
}

type CropsGrowInfo struct {
	//工序唯一ID
	CropsGrowId string `json:"crops_grow_id"`
	//物料ID
	CropsBakId string `json:"crops_bak_id"`
	//记录时间
	RecordTime string `json:"record_time"`
	//工序图片URL
	CropsGrowPhotoUrl string `json:"crops_grow_photo_url"`
	//施加电流（A）
	Temperature string `json:"temperature"`
	//工序名称
	GrowStatus string `json:"grow_status"`
	//短路阻抗实测值（%）
	WaterContent string `json:"water_content"`
	//负载损耗实测值（W）
	IlluminationStatus string `json:"illumination_status"`
	//备注
	Remarks string `json:"remarks"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	//根据物料ID查询物料
	if function == "queryCropsById" {
		return s.queryCropsById(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createCrops" {
		return s.createCrops(APIstub, args)
	} else if function == "queryCropsProcessByCropsId" {
		return s.queryCropsProcessByCropsId(APIstub, args)
	} else if function == "recordCropsGrow" {
		return s.recordCropsGrow(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

/**
 * 根据物料ID查询物料信息
 */
func (s *SmartContract) queryCropsById(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	cropsAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(cropsAsBytes)
}

//根据crops_id溯源所有记录过程
func (s *SmartContract) queryCropsProcessByCropsId(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	CropsBakId := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"crops_bak_id\":{\"$eq\":\"%s\"}}}", CropsBakId)
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

/**
 * 初始化账本
 */
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	crops := []Crops{
		Crops{CropsId: "steak_liu_first_record",
			CropsName:            "cherry",
			Address:              "ChengDu",
			RegisterTime:         "2022.6.1",
			Year:                 "2022",
			FarmerName:           "cloud",
			FarmerID:             "2319492349",
			FertilizerName:       "电磁线",
			PlatMode:             "是是是",
			BaggingStatus:        "套袋",
			GrowSeedlingsCycle:   "30天",
			IrrigationCycle:      "4天",
			ApplyFertilizerCycle: "无",
			WeedCycle:            "5天",
			Remarks:              "非常不错"},
	}
	i := 0
	for i < len(crops) {
		fmt.Println("i is ", i)
		cropsAsBytes, _ := json.Marshal(crops[i])
		APIstub.PutState(crops[i].CropsId, cropsAsBytes)
		fmt.Println("Added", crops[i])
		i = i + 1
	}
	return shim.Success(nil)
}

/**
 *记录物料工序
 */
func (s *SmartContract) recordCropsGrow(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	var cropsGrowInfo = CropsGrowInfo{CropsGrowId: args[1], CropsBakId: args[2], RecordTime: args[3], CropsGrowPhotoUrl: args[4], Temperature: args[5], GrowStatus: args[6], WaterContent: args[7], IlluminationStatus: args[8], Remarks: args[9]}
	cropsGrowInfoAsBytes, _ := json.Marshal(cropsGrowInfo)
	APIstub.PutState(args[0], cropsGrowInfoAsBytes)
	return shim.Success(nil)
}

/**
 *添加物料
 */
func (s *SmartContract) createCrops(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 16 {
		return shim.Error("Incorrect number of arguments. Expecting 15")
	}
	var crops = Crops{CropsId: args[1], CropsName: args[2], Address: args[3], FarmerName: args[4], FertilizerName: args[5], PlatMode: args[6], BaggingStatus: args[7], GrowSeedlingsCycle: args[8], IrrigationCycle: args[9], ApplyFertilizerCycle: args[10], WeedCycle: args[11], Remarks: args[12], RegisterTime: args[13], Year: args[14], FarmerTel: args[15]}
	cropsAsBytes, _ := json.Marshal(crops)
	APIstub.PutState(args[0], cropsAsBytes)
	return shim.Success(nil)
}

func (s *SmartContract) queryAllCropsByFarmerID(APIstub shim.ChaincodeStubInterface, args string) sc.Response {

	startKey := "CAR0"
	endKey := "CAR999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}
