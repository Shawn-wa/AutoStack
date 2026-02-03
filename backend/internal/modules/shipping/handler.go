package shipping

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"autostack/pkg/response"
)

// ========== 运费模板处理 ==========

// ListTemplates 获取运费模板列表
func ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	templates, total, err := service.ListTemplates(page, pageSize, keyword, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取模板列表失败")
		return
	}

	var list []TemplateResponse
	for _, t := range templates {
		list = append(list, TemplateResponse{
			ID:          t.ID,
			Name:        t.Name,
			Carrier:     t.Carrier,
			FromRegion:  t.FromRegion,
			Description: t.Description,
			Status:      t.Status,
			RuleCount:   len(t.Rules),
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", TemplateListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ListAllTemplates 获取所有启用的运费模板（用于下拉选择）
func ListAllTemplates(c *gin.Context) {
	templates, err := service.ListAllTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取模板列表失败")
		return
	}

	var list []TemplateOptionResponse
	for _, t := range templates {
		list = append(list, TemplateOptionResponse{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	response.Success(c, http.StatusOK, "获取成功", list)
}

// GetTemplate 获取运费模板详情
func GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的模板ID")
		return
	}

	template, err := service.GetTemplate(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "模板不存在")
		return
	}

	var rules []TemplateRuleResponse
	for _, r := range template.Rules {
		rules = append(rules, TemplateRuleResponse{
			ID:              r.ID,
			TemplateID:      r.TemplateID,
			ToRegion:        r.ToRegion,
			MinWeight:       r.MinWeight,
			MaxWeight:       r.MaxWeight,
			FirstWeight:     r.FirstWeight,
			FirstPrice:      r.FirstPrice,
			AdditionalUnit:  r.AdditionalUnit,
			AdditionalPrice: r.AdditionalPrice,
			Currency:        r.Currency,
			EstimatedDays:   r.EstimatedDays,
			CreatedAt:       r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Carrier:     template.Carrier,
		FromRegion:  template.FromRegion,
		Description: template.Description,
		Status:      template.Status,
		RuleCount:   len(rules),
		Rules:       rules,
		CreatedAt:   template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   template.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// CreateTemplate 创建运费模板
func CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	template, err := service.CreateTemplate(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建模板失败")
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Carrier:     template.Carrier,
		FromRegion:  template.FromRegion,
		Description: template.Description,
		Status:      template.Status,
		CreatedAt:   template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   template.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UpdateTemplate 更新运费模板
func UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的模板ID")
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.UpdateTemplate(uint(id), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新模板失败")
		return
	}

	response.Success(c, http.StatusOK, "更新成功", nil)
}

// DeleteTemplate 删除运费模板
func DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的模板ID")
		return
	}

	if err := service.DeleteTemplate(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除模板失败")
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}

// ========== 运费规则处理 ==========

// GetTemplateRules 获取模板的所有规则
func GetTemplateRules(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的模板ID")
		return
	}

	rules, err := service.GetRulesByTemplateID(uint(templateID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取规则列表失败")
		return
	}

	var list []TemplateRuleResponse
	for _, r := range rules {
		list = append(list, TemplateRuleResponse{
			ID:              r.ID,
			TemplateID:      r.TemplateID,
			ToRegion:        r.ToRegion,
			MinWeight:       r.MinWeight,
			MaxWeight:       r.MaxWeight,
			FirstWeight:     r.FirstWeight,
			FirstPrice:      r.FirstPrice,
			AdditionalUnit:  r.AdditionalUnit,
			AdditionalPrice: r.AdditionalPrice,
			Currency:        r.Currency,
			EstimatedDays:   r.EstimatedDays,
			CreatedAt:       r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, http.StatusOK, "获取成功", list)
}

// CreateRule 创建运费规则
func CreateRule(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的模板ID")
		return
	}

	var req CreateTemplateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	rule, err := service.CreateRule(uint(templateID), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "创建成功", TemplateRuleResponse{
		ID:              rule.ID,
		TemplateID:      rule.TemplateID,
		ToRegion:        rule.ToRegion,
		MinWeight:       rule.MinWeight,
		MaxWeight:       rule.MaxWeight,
		FirstWeight:     rule.FirstWeight,
		FirstPrice:      rule.FirstPrice,
		AdditionalUnit:  rule.AdditionalUnit,
		AdditionalPrice: rule.AdditionalPrice,
		Currency:        rule.Currency,
		EstimatedDays:   rule.EstimatedDays,
		CreatedAt:       rule.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UpdateRule 更新运费规则
func UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ruleId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的规则ID")
		return
	}

	var req UpdateTemplateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.UpdateRule(uint(id), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新规则失败")
		return
	}

	response.Success(c, http.StatusOK, "更新成功", nil)
}

// DeleteRule 删除运费规则
func DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ruleId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的规则ID")
		return
	}

	if err := service.DeleteRule(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除规则失败")
		return
	}

	response.Success(c, http.StatusOK, "删除成功", nil)
}

// ========== 运费计算处理 ==========

// CalculateShipping 计算运费
func CalculateShippingHandler(c *gin.Context) {
	var req CalculateShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := service.CalculateShipping(req.TemplateID, req.ToRegion, req.Weight)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "计算成功", result)
}

// BatchCalculateShipping 批量计算运费
func BatchCalculateShippingHandler(c *gin.Context) {
	var req BatchCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	results, err := service.BatchCalculateShipping(req.Items)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "计算失败")
		return
	}

	response.Success(c, http.StatusOK, "计算成功", BatchCalculateResponse{Results: results})
}

// ========== 本地产品运费模版关联处理 ==========

// BindProductShippingTemplate 绑定本地产品运费模版
func BindProductShippingTemplate(c *gin.Context) {
	var req BindProductShippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	pst, err := service.BindProductShippingTemplate(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "绑定成功", ConvertProductShippingTemplateToResponse(pst))
}

// UnbindProductShippingTemplate 解绑本地产品运费模版
func UnbindProductShippingTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := service.UnbindProductShippingTemplate(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑失败")
		return
	}

	response.Success(c, http.StatusOK, "解绑成功", nil)
}

// GetProductShippingTemplates 获取本地产品的运费模版列表
func GetProductShippingTemplates(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的产品ID")
		return
	}

	list, err := service.GetProductShippingTemplates(uint(productID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取失败")
		return
	}

	var respList []ProductShippingTemplateResponse
	for _, pst := range list {
		respList = append(respList, *ConvertProductShippingTemplateToResponse(&pst))
	}

	response.Success(c, http.StatusOK, "获取成功", respList)
}

// SetProductDefaultShippingTemplate 设置本地产品的默认运费模版
func SetProductDefaultShippingTemplate(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的产品ID")
		return
	}

	var req SetDefaultShippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.SetProductDefaultShippingTemplate(uint(productID), req.ShippingTemplateID); err != nil {
		response.Error(c, http.StatusInternalServerError, "设置失败")
		return
	}

	response.Success(c, http.StatusOK, "设置成功", nil)
}

// ========== 平台产品运费模版关联处理 ==========

// BindPlatformProductShippingTemplate 绑定平台产品运费模版
func BindPlatformProductShippingTemplate(c *gin.Context) {
	var req BindPlatformProductShippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	ppst, err := service.BindPlatformProductShippingTemplate(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "绑定成功", ConvertPlatformProductShippingTemplateToResponse(ppst))
}

// UnbindPlatformProductShippingTemplate 解绑平台产品运费模版
func UnbindPlatformProductShippingTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := service.UnbindPlatformProductShippingTemplate(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑失败")
		return
	}

	response.Success(c, http.StatusOK, "解绑成功", nil)
}

// GetPlatformProductShippingTemplates 获取平台产品的运费模版列表
func GetPlatformProductShippingTemplates(c *gin.Context) {
	platformProductID, err := strconv.ParseUint(c.Param("platformProductId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的平台产品ID")
		return
	}

	list, err := service.GetPlatformProductShippingTemplates(uint(platformProductID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取失败")
		return
	}

	var respList []PlatformProductShippingTemplateResponse
	for _, ppst := range list {
		respList = append(respList, *ConvertPlatformProductShippingTemplateToResponse(&ppst))
	}

	response.Success(c, http.StatusOK, "获取成功", respList)
}

// SetPlatformProductDefaultShippingTemplate 设置平台产品的默认运费模版
func SetPlatformProductDefaultShippingTemplate(c *gin.Context) {
	platformProductID, err := strconv.ParseUint(c.Param("platformProductId"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的平台产品ID")
		return
	}

	var req SetDefaultShippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.SetPlatformProductDefaultShippingTemplate(uint(platformProductID), req.ShippingTemplateID); err != nil {
		response.Error(c, http.StatusInternalServerError, "设置失败")
		return
	}

	response.Success(c, http.StatusOK, "设置成功", nil)
}
