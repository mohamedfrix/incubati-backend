# 📋 **HTTP Test Suite**

*Comprehensive API Testing for Application Service*  
*Date: July 3, 2025*

---

## 📁 **File Organization**

### **Core Test Files**
- `01_health.http` - Health check and utility endpoints
- `02_applications_crud.http` - Basic CRUD operations for applications
- `03_applications_internship.http` - Internship-specific application tests
- `04_applications_incubation.http` - Incubation-specific application tests
- `05_applications_pfe.http` - PFE-specific application tests
- `06_applications_status.http` - Status management and transitions
- `07_applications_query.http` - Query, filtering, and pagination tests
- `08_documents.http` - Document upload/download operations
- `09_error_cases.http` - Error handling and edge cases
- `10_integration_flows.http` - End-to-end integration scenarios

### **Support Files**
- `variables.http` - Environment variables and common values
- `test_files/` - Sample files for document upload testing

---

## 🚀 **Quick Start**

### **Prerequisites**
1. Application service running on `http://localhost:8080`
2. PostgreSQL database with migrations applied
3. Minio server running on `http://localhost:9000`
4. HTTP client (VS Code REST Client, Postman, etc.)

### **Testing Order**
1. Run `01_health.http` to verify service is running
2. Test basic CRUD with `02_applications_crud.http`
3. Test specific application types (`03_`, `04_`, `05_`)
4. Test advanced features (`06_`, `07_`, `08_`)
5. Verify error handling with `09_error_cases.http`
6. Run integration scenarios with `10_integration_flows.http`

### **Environment Configuration**
Edit `variables.http` to match your environment:
- `BASE_URL` - API base URL (default: http://localhost:8080)
- `USER_ID` - Test user ID
- File paths for document uploads

---

## 📊 **Test Coverage**

### **Functional Tests**
- ✅ All CRUD operations
- ✅ Type-specific data handling
- ✅ Status transitions
- ✅ Document management
- ✅ Query and filtering
- ✅ Pagination

### **Edge Cases**
- ✅ Invalid input validation
- ✅ Resource not found scenarios
- ✅ Status transition violations
- ✅ File upload edge cases
- ✅ Boundary value testing

### **Integration Tests**
- ✅ Complete application lifecycle
- ✅ Document upload/download flow
- ✅ Multi-step business processes

---

## 🎯 **Usage Tips**

### **VS Code REST Client**
1. Install "REST Client" extension
2. Open any `.http` file
3. Click "Send Request" above each request
4. View responses in the output panel

### **Variable Substitution**
- Use `{{variable}}` syntax for dynamic values
- Variables are defined in `variables.http`
- Some variables are auto-generated (like IDs from responses)

### **Authentication**
- Currently using basic user_id parameter
- Ready for JWT token integration when implemented

---

*These HTTP tests provide comprehensive coverage of all API endpoints with realistic test scenarios and edge cases.*
