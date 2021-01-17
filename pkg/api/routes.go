package api

import (
	"net/http"
	"next-terminal/pkg/global"
	"next-terminal/pkg/model"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const Token = "X-Auth-Token"

func SetupRoutes() *echo.Echo {

	e := echo.New()
	e.HideBanner = true

	e.File("/", "web/build/index.html")
	e.File("/logo.svg", "web/build/logo.svg")
	e.File("/favicon.ico", "web/build/favicon.ico")
	e.Static("/static", "web/build/static")

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(ErrorHandler)
	e.Use(Auth)

	e.POST("/login", LoginEndpoint)
	e.POST("/loginWithTotp", loginWithTotpEndpoint)

	e.GET("/tunnel", TunEndpoint)
	e.GET("/ssh", SSHEndpoint)

	e.POST("/logout", LogoutEndpoint)
	e.POST("/change-password", ChangePasswordEndpoint)
	e.POST("/reset-totp", ResetTOTPEndpoint)
	e.POST("/confirm-totp", ConfirmTOTPEndpoint)
	e.GET("/info", InfoEndpoint)

	users := e.Group("/users")
	{
		users.POST("", UserCreateEndpoint)
		users.GET("/paging", UserPagingEndpoint)
		users.PUT("/:id", UserUpdateEndpoint)
		users.DELETE("/:id", UserDeleteEndpoint)
		users.GET("/:id", UserGetEndpoint)
	}

	userGroups := e.Group("/user-groups")
	{
		userGroups.POST("", UserGroupCreateEndpoint)
		userGroups.GET("/paging", UserGroupPagingEndpoint)
		userGroups.PUT("/:id", UserGroupUpdateEndpoint)
		userGroups.DELETE("/:id", UserGroupDeleteEndpoint)
		userGroups.GET("/:id", UserGroupGetEndpoint)
		//userGroups.POST("/:id/members", UserGroupAddMembersEndpoint)
		//userGroups.DELETE("/:id/members/:memberId", UserGroupDelMembersEndpoint)
	}

	assets := e.Group("/assets", Auth)
	{
		assets.GET("", AssetAllEndpoint)
		assets.POST("", AssetCreateEndpoint)
		assets.GET("/paging", AssetPagingEndpoint)
		assets.POST("/:id/tcping", AssetTcpingEndpoint)
		assets.PUT("/:id", AssetUpdateEndpoint)
		assets.DELETE("/:id", AssetDeleteEndpoint)
		assets.GET("/:id", AssetGetEndpoint)
		assets.POST("/:id/change-owner", AssetChangeOwnerEndpoint)
	}

	e.GET("/tags", AssetTagsEndpoint)

	commands := e.Group("/commands")
	{
		commands.GET("/paging", CommandPagingEndpoint)
		commands.POST("", CommandCreateEndpoint)
		commands.PUT("/:id", CommandUpdateEndpoint)
		commands.DELETE("/:id", CommandDeleteEndpoint)
		commands.GET("/:id", CommandGetEndpoint)
		commands.POST("/:id/change-owner", CommandChangeOwnerEndpoint)
	}

	credentials := e.Group("/credentials")
	{
		credentials.GET("", CredentialAllEndpoint)
		credentials.GET("/paging", CredentialPagingEndpoint)
		credentials.POST("", CredentialCreateEndpoint)
		credentials.PUT("/:id", CredentialUpdateEndpoint)
		credentials.DELETE("/:id", CredentialDeleteEndpoint)
		credentials.GET("/:id", CredentialGetEndpoint)
		credentials.POST("/:id/change-owner", CredentialChangeOwnerEndpoint)
	}

	sessions := e.Group("/sessions")
	{
		sessions.POST("", SessionCreateEndpoint)
		sessions.GET("/paging", SessionPagingEndpoint)
		sessions.POST("/:id/content", SessionContentEndpoint)
		sessions.POST("/:id/discontent", SessionDiscontentEndpoint)
		sessions.POST("/:id/resize", SessionResizeEndpoint)
		sessions.POST("/:id/upload", SessionUploadEndpoint)
		sessions.GET("/:id/download", SessionDownloadEndpoint)
		sessions.GET("/:id/ls", SessionLsEndpoint)
		sessions.POST("/:id/mkdir", SessionMkDirEndpoint)
		sessions.DELETE("/:id/rmdir", SessionRmDirEndpoint)
		sessions.DELETE("/:id/rm", SessionRmEndpoint)
		sessions.DELETE("/:id", SessionDeleteEndpoint)
		sessions.GET("/:id/recording", SessionRecordingEndpoint)
		sessions.GET("/:id", SessionGetEndpoint)
	}

	resources := e.Group("/resources")
	{
		resources.GET("/:id/assign", ResourceGetAssignEndPoint)
		resources.POST("/:id/assign", ResourceOverwriteAssignEndPoint)
		resources.POST("/remove", ResourceRemoveByUserIdAssignEndPoint)
		resources.POST("/add", ResourceAddByUserIdAssignEndPoint)
	}

	e.GET("/properties", PropertyGetEndpoint)
	e.PUT("/properties", PropertyUpdateEndpoint)

	e.GET("/overview/counter", OverviewCounterEndPoint)
	e.GET("/overview/sessions", OverviewSessionPoint)

	return e
}

type H map[string]interface{}

func Fail(c echo.Context, code int, message string) error {
	return c.JSON(200, H{
		"code":    code,
		"message": message,
	})
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(200, H{
		"code":    1,
		"message": "success",
		"data":    data,
	})
}

func NotFound(c echo.Context, message string) error {
	return c.JSON(200, H{
		"code":    -1,
		"message": message,
	})
}

func GetToken(c echo.Context) string {
	token := c.Request().Header.Get(Token)
	if len(token) > 0 {
		return token
	}
	return c.QueryParam(Token)
}

func GetCurrentAccount(c echo.Context) (model.User, bool) {
	token := GetToken(c)
	get, b := global.Cache.Get(token)
	if b {
		return get.(Authorization).User, true
	}
	return model.User{}, false
}

func HasPermission(c echo.Context, owner string) bool {
	// 检测是否为创建者
	account, found := GetCurrentAccount(c)
	if !found {
		return false
	}
	if model.TypeAdmin == account.Type {
		return true
	}

	if owner == account.ID {
		return true
	}
	return false
}
