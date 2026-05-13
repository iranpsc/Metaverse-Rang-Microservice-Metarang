from rest_framework.routers import DefaultRouter
from . import views
from django.urls import path, include



app_name = 'api-v1'


router = DefaultRouter()

router.register('article', views.ArticleModelViewSet, basename='article')
router.register('category', views.CategoryModelViewSet, basename='category')
router.register('subcategory', views.SubCategoryModelViewSet, basename='subcategory')
router.register('tag', views.TagModelViewSet, basename='tag')
router.register('comment', views.CommentModelViewSet, basename='comment')
router.register('reply', views.ReplyModelViewSet, basename='reply')
router.register('likedislike', views.LikeDislikeModelViewSet, basename='like_dislike')



urlpatterns = [
    path('', include(router.urls)),
]