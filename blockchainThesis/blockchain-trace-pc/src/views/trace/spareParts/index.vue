<template>
	<div class="app-container">

		<el-row :gutter="10" class="mb8">
			<el-col :span="1.5"><el-button type="primary" icon="el-icon-plus" size="mini" @click="handleAdd">添加物料</el-button></el-col>
			<right-toolbar :showSearch.sync="showSearch" @queryTable="getList"></right-toolbar>
		</el-row>

		<el-table v-loading="loading" :data="cropsList" @selection-change="handleSelectionChange">
			<el-table-column type="selection" width="55" align="center" />
			<el-table-column label="物料编码" prop="cropsId" />
			<el-table-column label="物料名称" prop="cropsName" />
			<el-table-column label="状态" prop="status">
				<template slot-scope="scope">
					<el-tag v-if="scope.row.status === 0">作业中</el-tag>
					<el-tag type="danger" v-if="scope.row.status === 1">已完成</el-tag>
				</template>
			</el-table-column>
			<el-table-column label="操作" align="center" class-name="small-padding fixed-width">
				<template slot-scope="scope">
					<el-button v-show="scope.row.status === 0" size="mini" type="text" @click="handleRecord(scope.row)">物料规格参数</el-button>
					<el-button size="mini" type="text" @click="cropsDetail(scope.row)">物料详情</el-button>
					<el-button size="mini" type="text" @click="cropsProcessDetail(scope.row)">工序详情</el-button>
					<el-button v-show="scope.row.status === 0" size="mini" type="text" @click="noticeTrasport(scope.row)">通知运输</el-button>
					<el-button size="mini" type="text" @click="logisticsTrace(scope.row)">物流追踪</el-button>
				</template>
			</el-table-column>
		</el-table>

		<pagination v-show="total > 0" :total="total" :page.sync="queryParams.pageNum" :limit.sync="queryParams.pageSize" @pagination="getList" />

		<!-- 记录物料工序 -->
		<el-dialog center title="记录工序规格参数" :visible.sync="growDialog" width="700px" append-to-body>
			<el-form ref="form" label-width="80px" :model="recordForm">
				<el-row>
					<el-col :span="24" align="center">{{ cropsName }}</el-col>
				</el-row>
				<el-divider></el-divider>
				<el-row>
					<el-col :span="12">
						<el-form-item label="施加电流（A）" prop="nickName"><el-input v-model="recordForm.temperature" placeholder="请输入施加电流（A）" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="工序名称" prop="nickName"><el-input v-model="recordForm.growStatus" placeholder="请输入工序名称" /></el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="短路阻抗实测值（%）" prop="nickName"><el-input v-model="recordForm.waterContent" placeholder="请输入短路阻抗实测值（%）" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="负载损耗实测值（W）" prop="nickName"><el-input v-model="recordForm.illuminationStatus" placeholder="请输入负载损耗实测值（W）" /></el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="备注" prop="nickName"><el-input v-model="recordForm.remarks" type="textarea" placeholder="请输入内容"></el-input></el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="图片:">
							<el-upload class="avatar-uploader" :on-change="getFile" :show-file-list="false" :auto-upload="false">
								<img v-if="imageUrl" :src="imageUrl" class="avatar" />
								<i v-else class="el-icon-plus avatar-uploader-icon"></i>
							</el-upload>
						</el-form-item>
					</el-col>
				</el-row>
			</el-form>
			<div slot="footer" class="dialog-footer">
				<el-button size="mini" type="primary" @click="recordCropsGrow">确 定</el-button>
				<el-button size="mini" @click="cancel">取 消</el-button>
			</div>
		</el-dialog>

		<!-- 添加物料 -->
		<el-dialog :title="title" :visible.sync="open" width="700px" append-to-body>
			<el-form ref="form" :model="form" :rules="rules" label-width="80px">
				<el-row>
					<el-col :span="12">
						<el-form-item label="物料名称" prop="nickName"><el-input v-model="form.cropsName" placeholder="请输入物料名称" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="物料类型" prop="nickName">
							<el-select v-model="form.cropsType" placeholder="请选择">
								<el-option v-for="dict in cropsTypeOptions" :key="dict.dictValue" :label="dict.dictLabel" :value="dict.dictValue"></el-option>
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="合同编号(国网经法)" prop="nickName"><el-input v-model="form.fertilizerName" placeholder="请输入合同编号(国网经法)" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="品类编码" prop="nickName">
							<el-select v-model="form.plantMode" placeholder="请选择">
								<el-option v-for="dict in plantModeOptions" :key="dict.dictValue" :label="dict.dictLabel" :value="dict.dictValue"></el-option>
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="排产计划编码" prop="nickName"><el-input v-model="form.year" placeholder="请输入排产计划编码" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="种类编码" prop="nickName">
							<el-select v-model="form.baggingStatus" placeholder="请选择">
								<el-option v-for="dict in beggingOptions" :key="dict.dictValue" :label="dict.dictLabel" :value="dict.dictValue"></el-option>
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="采购订单周期" prop="nickName"><el-input v-model="form.growSeedlingsCycle" placeholder="请输入：几周" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="销售订单周期" prop="nickName"><el-input v-model="form.irrigationCycle" placeholder="请输入：几周" /></el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="排产计划周期" prop="nickName"><el-input v-model="form.applyFertilizerCycle" placeholder="请输入：几周" /></el-form-item>
					</el-col>
					<el-col :span="12">
						<el-form-item label="生产订单周期" prop="nickName"><el-input v-model="form.weedCycle" placeholder="请输入：几周" /></el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="12">
						<el-form-item label="计量单位" prop="nickName"><el-input v-model="form.address" placeholder="请输入计量单位" /></el-form-item>
					</el-col>
				</el-row>
				<el-form-item label="备注"><el-input v-model="form.remarks" type="textarea" placeholder="请输入内容"></el-input></el-form-item>
			</el-form>
			<div slot="footer" class="dialog-footer">
				<el-button type="primary" @click="createCrops">确 定</el-button>
				<el-button @click="cancel">取 消</el-button>
			</div>
		</el-dialog>

		<!-- 物料详情 -->
		<el-dialog center title="订单链上数据详情" :visible.sync="cropsDetaiDialog" width="700px" append-to-body>
			<el-row>
				<el-col :span="12">物料ID：{{ cropsDetails.crops_id }}</el-col>
				<el-col :span="12">物料名称：{{ cropsDetails.crops_name }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
				<el-col :span="12">计量单位：{{ cropsDetails.address }}</el-col>
				<el-col :span="12">原材料厂商：{{ cropsDetails.farmer_name }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
        <el-col :span="12">登记时间：{{ cropsDetails.register_time }}</el-col>
				<el-col :span="12">品类编码：{{ cropsDetails.plant_mode }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
				<el-col :span="12">合同编号(国网经法)：{{ cropsDetails.fertilizer_name }}</el-col>
				<el-col :span="12">种类编码：{{ cropsDetails.bagging_status }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
				<el-col :span="12">原材料厂商电话：{{ cropsDetails.farmer_tel }}</el-col>
				<el-col :span="12">排产计划编码：{{ cropsDetails.year }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
				<el-col :span="12">采购订单周期：{{ cropsDetails.grow_seedlings_cycle }}</el-col>
				<el-col :span="12">销售订单周期：{{ cropsDetails.irrigation_cycle }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
        <el-col :span="12">排产计划周期：{{ cropsDetails.apply_fertilizer_cycle }}</el-col>
				<el-col :span="12">生产订单周期：{{ cropsDetails.weed_cycle }}</el-col>
			</el-row>
			<el-divider></el-divider>
			<el-row>
				<el-col :span="12">备注：{{ cropsDetails.remarks }}</el-col>
			</el-row>
		</el-dialog>

		<!-- 过程详情 -->
		<el-drawer size="80%" :visible.sync="cropsProcessDetaiDialog" :show-close="false" :with-header="false">
			<el-divider content-position="left">工序链上过程详情</el-divider>
			<el-table v-loading="loading" :data="cropsProcessDetailsArray">
				<el-table-column label="过程ID" prop="crops_grow_id" />
				<el-table-column label="工程项目ID" prop="crops_bak_id" />
				<el-table-column label="工序名称" prop="grow_status" />
				<el-table-column label="负载损耗实测值（W）" prop="illumination_status" />
				<el-table-column label="记录时间" prop="record_time" />
				<el-table-column label="施加电流（A）" prop="temperature" />
				<el-table-column label="短路阻抗实测值（%）" prop="water_content" />
				<el-table-column label="备注" prop="remarks" />
				<el-table-column label="图片" class="demo-image__preview">
					<template slot-scope="scope">
						<el-image :preview-src-list="scope.row.crops_grow_photo_url" :src="scope.row.crops_grow_photo_url" style="width: 100px;height: 80px"></el-image>
					</template>
				</el-table-column>
			</el-table>
		</el-drawer>

		<el-dialog center title="联系运输" :visible.sync="noticeDetaiDialog" width="500px" append-to-body>
			<el-form ref="form" :model="trasportForm" label-width="80px">
				<el-row>
					<el-col :span="24">
						<el-form-item label="选择司机" prop="nickName">
							<el-select v-model="trasportForm.userName" placeholder="请选择">
								<el-option v-for="user in driverList" :key="user.userName" :label="user.nickName" :value="user.userName"></el-option>
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="24">
						<el-form-item label="原料厂商" prop="nickName">
							<el-select v-model="trasportForm.deptId" placeholder="请选择">
								<el-option v-for="dept in factoryList" :key="dept.deptId" :label="dept.deptName" :value="dept.deptId"></el-option>
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row>
					<el-col :span="24">
						<el-form-item label="备注">
							<el-input v-model="trasportForm.remarks" type="textarea" placeholder="请输入内容"></el-input>
						</el-form-item>
					</el-col>
				</el-row>
			</el-form>
			<div slot="footer" class="dialog-footer">
				<el-button type="primary" @click="addNoticeTrasport">确 定</el-button>
				<el-button @click="cancel">取 消</el-button>
			</div>
		</el-dialog>

		<!--轨迹弹出框-->
		<el-drawer size="80%" :visible.sync="playTrackView" :show-close="false" :with-header="false">
			<el-divider content-position="left">物流追踪</el-divider>
			<!-- <el-container style="height: 600px; border: 1px solid #eee">
				<el-main>
					<div class="map-page" style="height: 100%"><div id="position"></div></div>
				</el-main>
			</el-container> -->
			<div id="container" class="map"></div>
		</el-drawer>
	</div>
</template>

<script>
import { listRole, getRole, delRole, addRole, updateRole, exportRole, dataScope, changeRoleStatus } from '@/api/system/role';
import { treeselect as menuTreeselect, roleMenuTreeselect } from '@/api/system/menu';
import { treeselect as deptTreeselect, roleDeptTreeselect } from '@/api/system/dept';
import {getFactoryByDeptId, addTransport, getAllDriverByDeptId, uplodImagesBase64, listCrops, saveCrops } from '../../../api/trace/spareParts';
import { formatDate } from '../../../utils/index.js';
export default {
	name: 'Role',

	data() {
		return {
			// 遮罩层
			loading: true,
			// 选中数组
			ids: [],
			// 非单个禁用
			single: true,
			// 非多个禁用
			multiple: true,
			// 显示搜索条件
			showSearch: true,
			// 总条数
			total: 0,
			// 角色表格数据
			cropsList: [],
			// 弹出层标题
			title: '',
			// 是否显示弹出层
			open: false,
			// 是否显示弹出层（数据权限）
			openDataScope: false,
			menuExpand: false,
			menuNodeAll: false,
			deptExpand: true,
			deptNodeAll: false,
			// 日期范围
			dateRange: [],
			// 状态数据字典
			statusOptions: ['在生产', '停止生产'],

			beggingOptions: [],
			plantModeOptions: [], //生产方式
			cropsTypeOptions: [], //物料类型

			growDialog: false, //物料工序记录弹出框
			cropsName: '',
			cropsDetaiDialog: false,

			cropsDetails: '',
			// 菜单列表
			menuOptions: [],
			// 部门列表
			deptOptions: [],
			// 查询参数
			queryParams: {
				pageNum: 1,
				pageSize: 10,
				roleName: undefined,
				roleKey: undefined,
				status: undefined
			},
			// 表单参数
			form: {},
			defaultProps: {
				children: 'children',
				label: 'label'
			},
			recordForm: {},
			imageUrl: '',
			imgBase64: [],
			cropsPhotoUrl: '',
			cropsInfo: '',
			growDialog: false,
			cropsProcessDetailsArray: [],
			cropsProcessDetaiDialog: false,
			noticeDetaiDialog: false,
			driverList: [],
			trasportForm: {
				cropsId: '',
				farmerUserName: '',
				farmerNickName: '',
				time: ''
			},
			trasportInfo: '',
			playTrackView: false,
			factoryList:[],
		};
	},
	created() {
		this.getCropsList();
		this.getDicts('crops_bagging').then(response => {
			this.beggingOptions = response.data;
		});
		this.getDicts('crops_plant_type').then(res => {
			this.plantModeOptions = res.data;
		});
		this.getDicts('crops_type').then(res => {
			this.cropsTypeOptions = res.data;
		});
	},
	methods: {
		/**
		 * 物流追踪
		 */
		logisticsTrace(row) {
		  this.$nextTick(() => {
        this.playTrackView = true;
        const lineArr = [
          [103.8878, 30.6986],
          [103.8879, 30.6987],
          [112.4809, 37.811],
          [112.5486, 37.8605],
        ];
        //创建地图
        var map = new AMap.Map('container', {
          resizeEnable: true,
          center: [109.7259, 38.2663],
          zoom: 9
        });
        //标记车辆
        var marker = new AMap.Marker({
          position: [109.7259, 38.2663],
          icon: 'https://webapi.amap.com/images/car.png',
          //坐标偏移
          offset: new AMap.Pixel(-26, -13),
          autoRotation: true,
          angle: -90,
          map: map
        });
        // 绘制轨迹路线
        var polyline = new AMap.Polyline({
          map: this.map,
          //这里替换自己的坐标数据
          path: lineArr,
          borderWeight: 2, // 线条宽度，默认为 1
          strokeColor: 'red', // 线条颜色
          lineJoin: 'round' // 折线拐点连接处样式
        });
        map.add(polyline);

        //调用方法开启动画
        marker.moveAlong(lineArr, 30000);
      })

		},

		addNoticeTrasport() {
			this.trasportForm.cropsId = this.trasportInfo.cropsId;
			this.trasportForm.farmerUserName = this.$store.getters.name;
			this.trasportForm.farmerNickName = this.$store.getters.nickName;
			this.trasportForm.time = new Date();
			this.trasportForm.driverId = this.trasportForm.userName;
			this.trasportForm.status = 0;
			this.trasportForm.farmerTel = this.$store.getters.phonenumber;
			this.trasportForm.farmerDept = this.$store.getters.deptName;
			this.trasportForm.factoryId = this.trasportForm.deptId
			addTransport(this.trasportForm)
				.then(res => {
					this.msgSuccess('联系运输成功');
					this.noticeDetaiDialog = false;
					this.getCropsList();
				})
				.catch(err => {
					this.msgError('联系运输失败，发生异常');
				});
		},

		/**
		 * @param {Object} row通知运输
		 */
		noticeTrasport(row) {
			this.trasportInfo = row;
			getAllDriverByDeptId(this.$store.getters.deptId)
				.then(res => {
					this.driverList = res.data;
				})
				.catch(err => {});
				//获取所有原料厂商
			getFactoryByDeptId(119)
				.then(res => {
					this.factoryList = res.data;
				})
				.catch(err => {});
			this.noticeDetaiDialog = true;
		},

		/**
		 * 工序过程详情
		 */
		cropsProcessDetail(row) {
			this.$httpBlock
				.get(this.$httpUrl + '/farmerapi/queryCropsProcessByCropsId?id=' + row.cropsId)
				.then(res => {
					const array = [];

					for (let i = 0; i < res.data.length; i++) {
						array.push(res.data[i].Record);
					}
					this.cropsProcessDetailsArray = array;
					this.cropsProcessDetaiDialog = true;
				})
				.catch(err => {});
		},

		recordCropsGrow() {
			const cropsGrowInfo = this.recordForm;
			const cropsGrowArray = [];
			const id = new this.$snowFlakeId().generate();
			cropsGrowArray.push(id);
			cropsGrowArray.push(id);
			cropsGrowArray.push(this.cropsInfo.cropsId);
			cropsGrowArray.push(formatDate(new Date()));
			cropsGrowArray.push(this.cropsPhotoUrl);
			cropsGrowArray.push(cropsGrowInfo.temperature);
			cropsGrowArray.push(cropsGrowInfo.growStatus);
			cropsGrowArray.push(cropsGrowInfo.waterContent);
			cropsGrowArray.push(cropsGrowInfo.illuminationStatus);
			cropsGrowArray.push(cropsGrowInfo.remarks);
			this.$httpBlock
				.post(this.$httpUrl + '/farmerapi/recordCropsGrow', { cropsGrowArray: cropsGrowArray })
				.then(res => {
					if (res.status === 200) {
						this.msgSuccess('数据上链成功');
						this.growDialog = false;
					}
				})
				.catch(err => {
					this.msgError('存储异常 ' + err);
				});
		},

		getFile(file) {
			this.imageUrl = URL.createObjectURL(file.raw);
			this.getBase64(file.raw).then(res => {
				this.imgBase64 = res;
				const datas = {
					imageBase64: this.imgBase64
				};
				uplodImagesBase64(datas)
					.then(res => {
						this.cropsPhotoUrl = res.msg;
						this.msgSuccess('图片上传成功');
					})
					.catch(err => {
						this.msgError('图片上传失败');
					});
			});
		},

		getBase64(file) {
			return new Promise(function(resolve, reject) {
				const reader = new FileReader();
				let imgResult = '';
				reader.readAsDataURL(file);
				reader.onload = function() {
					imgResult = reader.result;
				};
				reader.onerror = function(error) {
					reject(error);
				};
				reader.onloadend = function() {
					resolve(imgResult);
				};
			});
		},

		createCrops() {
			const crops = this.form;
			const cropsArray = [];
			cropsArray.push(new this.$snowFlakeId().generate());
			cropsArray.push(new this.$snowFlakeId().generate());
			cropsArray.push(this.form.cropsName);
			cropsArray.push(this.form.address);
			cropsArray.push(this.$store.getters.nickName);
			cropsArray.push(this.form.fertilizerName);
			cropsArray.push(this.form.plantMode);
			cropsArray.push(this.form.baggingStatus);
			cropsArray.push(this.form.growSeedlingsCycle);
			cropsArray.push(this.form.irrigationCycle);
			cropsArray.push(this.form.applyFertilizerCycle);
			cropsArray.push(this.form.weedCycle);
			cropsArray.push(this.form.remarks);
			cropsArray.push(formatDate(new Date()));
			cropsArray.push(this.form.year);
			cropsArray.push(this.$store.getters.phonenumber);
			this.$httpBlock
				.post(this.$httpUrl + '/farmerapi/createCrops', { cropsArray: cropsArray })
				.then(res => {
					if (res.status === 200) {
						this.msgSuccess('数据上链成功');
						const cropsUser = {
							username: this.$store.getters.name,
							cropsId: cropsArray[0],
							cropsName: cropsArray[2],
							status: 0
						};
						saveCrops(cropsUser)
							.then(res => {
								this.getCropsList()
							})
							.catch(err => {
								this.msgError('id存储异常 ' + err);
							});
						this.getCropsList();
						this.open = false;
					}
				})
				.catch(err => {
					this.msgError('存储异常 ' + err);
				});
		},

		handleRecord(row) {
			this.cropsName = row.cropsName;
			this.cropsInfo = row;
			this.growDialog = true;
		},

		/** 查询物料列表 */
		getCropsList() {
			this.loading = true;
			listCrops(this.$store.getters.name).then(response => {
				this.cropsList = response.data;
				this.total = response.total;
				this.loading = false;
			});
		},

		/**
		 * 物料详情
		 */
		cropsDetail(row) {
			this.$httpBlock
				.get(this.$httpUrl + '/farmerapi/queryCropsById?id=' + row.cropsId)
				.then(res => {
					this.cropsDetails = res.data;
					this.cropsDetaiDialog = true;
				})
				.catch(err => {});
		},

		/** 查询菜单树结构 */
		getMenuTreeselect() {
			menuTreeselect().then(response => {
				this.menuOptions = response.data;
			});
		},
		/** 查询部门树结构 */
		getDeptTreeselect() {
			deptTreeselect().then(response => {
				this.deptOptions = response.data;
			});
		},
		// 所有菜单节点数据
		getMenuAllCheckedKeys() {
			// 目前被选中的菜单节点
			let checkedKeys = this.$refs.menu.getHalfCheckedKeys();
			// 半选中的菜单节点
			let halfCheckedKeys = this.$refs.menu.getCheckedKeys();
			checkedKeys.unshift.apply(checkedKeys, halfCheckedKeys);
			return checkedKeys;
		},
		// 所有部门节点数据
		getDeptAllCheckedKeys() {
			// 目前被选中的部门节点
			let checkedKeys = this.$refs.dept.getHalfCheckedKeys();
			// 半选中的部门节点
			let halfCheckedKeys = this.$refs.dept.getCheckedKeys();
			checkedKeys.unshift.apply(checkedKeys, halfCheckedKeys);
			return checkedKeys;
		},
		/** 根据角色ID查询菜单树结构 */
		getRoleMenuTreeselect(roleId) {
			return roleMenuTreeselect(roleId).then(response => {
				this.menuOptions = response.menus;
				return response;
			});
		},
		/** 根据角色ID查询部门树结构 */
		getRoleDeptTreeselect(roleId) {
			return roleDeptTreeselect(roleId).then(response => {
				this.deptOptions = response.depts;
				return response;
			});
		},

		// 取消按钮
		cancel() {
			this.open = false;
			this.reset();
		},
		// 取消按钮（数据权限）
		cancelDataScope() {
			this.openDataScope = false;
			this.reset();
		},
		// 表单重置
		reset() {
			if (this.$refs.menu != undefined) {
				this.$refs.menu.setCheckedKeys([]);
			}
			(this.menuExpand = false),
				(this.menuNodeAll = false),
				(this.deptExpand = true),
				(this.deptNodeAll = false),
				(this.form = {
					roleId: undefined,
					roleName: undefined,
					roleKey: undefined,
					roleSort: 0,
					status: '0',
					menuIds: [],
					deptIds: [],
					menuCheckStrictly: true,
					deptCheckStrictly: true,
					remark: undefined
				});
			this.resetForm('form');
		},
		/** 搜索按钮操作 */
		handleQuery() {
			this.queryParams.pageNum = 1;
			this.getList();
		},
		/** 重置按钮操作 */
		resetQuery() {
			this.dateRange = [];
			this.resetForm('queryForm');
			this.handleQuery();
		},
		// 多选框选中数据
		handleSelectionChange(selection) {
			this.ids = selection.map(item => item.roleId);
			this.single = selection.length != 1;
			this.multiple = !selection.length;
		},
		// 树权限（展开/折叠）
		handleCheckedTreeExpand(value, type) {
			if (type == 'menu') {
				let treeList = this.menuOptions;
				for (let i = 0; i < treeList.length; i++) {
					this.$refs.menu.store.nodesMap[treeList[i].id].expanded = value;
				}
			} else if (type == 'dept') {
				let treeList = this.deptOptions;
				for (let i = 0; i < treeList.length; i++) {
					this.$refs.dept.store.nodesMap[treeList[i].id].expanded = value;
				}
			}
		},
		// 树权限（全选/全不选）
		handleCheckedTreeNodeAll(value, type) {
			if (type == 'menu') {
				this.$refs.menu.setCheckedNodes(value ? this.menuOptions : []);
			} else if (type == 'dept') {
				this.$refs.dept.setCheckedNodes(value ? this.deptOptions : []);
			}
		},
		// 树权限（父子联动）
		handleCheckedTreeConnect(value, type) {
			if (type == 'menu') {
				this.form.menuCheckStrictly = value ? true : false;
			} else if (type == 'dept') {
				this.form.deptCheckStrictly = value ? true : false;
			}
		},
		/** 新增按钮操作 */
		handleAdd() {
			this.reset();
			this.getMenuTreeselect();
			this.open = true;
			this.title = '记录工序';
		},
		/** 修改按钮操作 */
		handleUpdate(row) {
			this.reset();
			const roleId = row.roleId || this.ids;
			const roleMenu = this.getRoleMenuTreeselect(roleId);
			getRole(roleId).then(response => {
				this.form = response.data;
				this.open = true;
				this.$nextTick(() => {
					roleMenu.then(res => {
						this.$refs.menu.setCheckedKeys(res.checkedKeys);
					});
				});
				this.title = '修改物料';
			});
		},
		/** 分配数据权限操作 */
		handleDataScope(row) {
			this.reset();
			const roleDeptTreeselect = this.getRoleDeptTreeselect(row.roleId);
			getRole(row.roleId).then(response => {
				this.form = response.data;
				this.openDataScope = true;
				this.$nextTick(() => {
					roleDeptTreeselect.then(res => {
						this.$refs.dept.setCheckedKeys(res.checkedKeys);
					});
				});
				this.title = '分配数据权限';
			});
		},
		/** 提交按钮 */
		submitForm: function() {
			// this.$refs["form"].validate(valid => {
			//     if (valid) {
			saveCrops(this.form).then(res => {
				if (res.code === 200) {
					this.msgSuccess('修改成功');
					this.open = false;
				}
			});
			// if (this.form.roleId != undefined) {
			//     this.form.menuIds = this.getMenuAllCheckedKeys();
			//     updateRole(this.form).then(response => {
			//         if (response.code === 200) {
			//             this.msgSuccess("修改成功");
			//             this.open = false;
			//             this.getList();
			//         }
			//     });
			// } else {
			//     this.form.menuIds = this.getMenuAllCheckedKeys();
			//     addRole(this.form).then(response => {
			//         if (response.code === 200) {
			//             this.msgSuccess("新增成功");
			//             this.open = false;
			//             this.getList();
			//         }
			//     });
			// }
			//     }
			// });
		},
		/** 提交按钮（数据权限） */
		submitDataScope: function() {
			if (this.form.roleId != undefined) {
				this.form.deptIds = this.getDeptAllCheckedKeys();
				dataScope(this.form).then(response => {
					if (response.code === 200) {
						this.msgSuccess('修改成功');
						this.openDataScope = false;
						this.getList();
					}
				});
			}
		},
		/** 删除按钮操作 */
		handleDelete(row) {
			const roleIds = row.roleId || this.ids;
			this.$confirm('是否确认删除角色编号为"' + roleIds + '"的数据项?', '警告', {
				confirmButtonText: '确定',
				cancelButtonText: '取消',
				type: 'warning'
			})
				.then(function() {
					return delRole(roleIds);
				})
				.then(() => {
					this.getList();
					this.msgSuccess('删除成功');
				})
				.catch(function() {});
		},
		/** 导出按钮操作 */
		handleExport() {
			const queryParams = this.queryParams;
			this.$confirm('是否确认导出所有角色数据项?', '警告', {
				confirmButtonText: '确定',
				cancelButtonText: '取消',
				type: 'warning'
			})
				.then(function() {
					return exportRole(queryParams);
				})
				.then(response => {
					this.download(response.msg);
				})
				.catch(function() {});
		}
	}
};
</script>
<style>
.avatar-uploader .el-upload {
	border: 1px dashed #d9d9d9;
	border-radius: 6px;
	cursor: pointer;
	position: relative;
	overflow: hidden;
}
.avatar-uploader .el-upload:hover {
	border-color: #409eff;
}
.avatar-uploader-icon {
	font-size: 28px;
	color: #8c939d;
	width: 178px;
	height: 178px;
	line-height: 178px;
	text-align: center;
}
.avatar {
	width: 178px;
	height: 178px;
	display: block;
}

.image {
	width: 80%;
	display: block;
}
/**一定要给宽高，否则不显示*/
#position {
	height: 100%;
}
/**修改点标记图片的css, 图片给宽高，否则不显示 */
.map-page /deep/ {
	.amap-icon {
		width: 20px;
	}
	.amap-marker-label {
		border: 1px solid #ccc;
		font-size: 16px;
		display: inline-block;
		padding: 5px;
	}
}

#panel {
	position: fixed;
	background-color: white;
	max-height: 90%;
	overflow-y: auto;
	top: 10px;
	right: 10px;
	width: 280px;
}
#panel .amap-call {
	background-color: #009cf9;
	border-top-left-radius: 4px;
	border-top-right-radius: 4px;
}
#panel .amap-lib-driving {
	border-bottom-left-radius: 4px;
	border-bottom-right-radius: 4px;
	overflow: hidden;
}
html,
body,
#container {
	width: 100%;
	height: 100%;
}

.map {
	width: 100%;
	height: 500px;
}
</style>
