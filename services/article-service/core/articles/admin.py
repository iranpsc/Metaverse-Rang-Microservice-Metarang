from django.contrib import admin
from articles.models import Article, Category, SubCategory, Tag, Comment, Reply, LikeDislike


class ArticlesAdmin(admin.ModelAdmin):
    list_display = ['identifier', 'slug', 'title', 'category', 'created_date', 'updated_date']




class TagsAdmin(admin.ModelAdmin):
    list_display = ['label', 'slug']




class SubCategoryAdmin(admin.ModelAdmin):
    list_display = ['category', 'name']



class CategoryAdmin(admin.ModelAdmin):
    list_display = ['name', 'category_slug']



class CommentsAdmin(admin.ModelAdmin):
    list_display = ['article', 'username', 'citizenId', 'created_date']




class RepliesAdmin(admin.ModelAdmin):
    list_display = ['comment', 'username', 'citizenId', 'created_date']



class LikeDislikeAdmin(admin.ModelAdmin):
    list_display = ['vote', 'username', 'citizenId', 'article', 'comment', 'reply']



admin.site.register(Article, ArticlesAdmin)
admin.site.register(Category, CategoryAdmin)
admin.site.register(SubCategory, SubCategoryAdmin)
admin.site.register(Tag, TagsAdmin)
admin.site.register(Comment, CommentsAdmin)
admin.site.register(Reply, RepliesAdmin)
admin.site.register(LikeDislike, LikeDislikeAdmin)