package pages

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// Bootstrap configures gin router according to a tree of PageGroups
// Next step should be to configure gin.Engine renderer using Pages.
func Bootstrap(r *gin.Engine, group *PageGroup) {
    if group.Index != "" {
        r.GET(group.Path, func(c *gin.Context) {
            c.HTML(http.StatusOK, group.Index, gin.H{})
        })
    }

    for _, page := range group.Pages {
        r.GET(page.GetPath(), func (c *gin.Context) {
            c.HTML(http.StatusOK, group.Index, gin.H{})    
        })
    }

    for _, child := range group.Groups {
        pgroup := r.Group(child.Path)

        setupRoutes(pgroup, child)
    }
}

func setupRoutes(r *gin.RouterGroup, group *PageGroup) {
    if group.HasIndex() {
        r.GET("", func(c *gin.Context) {
            c.HTML(http.StatusOK, group.Index, gin.H{})
        })
    }

    for _, page := range group.Pages {
        r.GET(page.GetPath(), func (c *gin.Context) {
            data := gin.H{}
            for _, param := range page.GetParams() {
                data[param] = c.Param(param)
            }
            c.HTML(http.StatusOK, page.AbsPath(), data)
        })
    }

    for _, child := range group.Groups {
        pgroup := r.Group(child.Path)

        setupRoutes(pgroup, child)
    }
}
