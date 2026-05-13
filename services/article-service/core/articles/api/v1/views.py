from rest_framework import viewsets, status
from . serializers import *
from ...models import Article, Category, SubCategory, Tag, Comment, Reply, LikeDislike
from rest_framework.filters import SearchFilter, OrderingFilter
from django_filters.rest_framework import DjangoFilterBackend 
from .paginations import ArticlePagination
from rest_framework.response import Response
from .permissions import IsOwnerOrReadOnly, IsAdminCitizenOrReadOnly, IsAuthenticatedRemote
from django.db.models import Count, Q
from django.db import IntegrityError
from rest_framework.exceptions import ValidationError, PermissionDenied
from django.db.models import Prefetch




class ArticleModelViewSet(viewsets.ModelViewSet):

    queryset = Article.objects.all()
    filter_backends = [DjangoFilterBackend, SearchFilter, OrderingFilter]
    permission_classes = [IsAdminCitizenOrReadOnly]
 
    filterset_fields = {'category':['exact', 'in']}
    search_fields = ['title', 'content', 'short_description']
    ordering_fields = ['created_date']
    pagination_class = ArticlePagination

    def get_serializer_class(self):

        if self.action == 'create':
            return ArticleCreateSerializer
        return ArticleListSerializer


    def get_queryset(self):
        comment_qs = Comment.objects.filter(approved=True).annotate(
            like_count=Count('likes', filter=Q(likes__vote=1)),
            dislike_count=Count('likes', filter=Q(likes__vote=-1))
        ).prefetch_related(
            Prefetch(
                'replies',
                queryset=Reply.objects.filter(approved=True).annotate(
                    like_count=Count('likes', filter=Q(likes__vote=1)),
                    dislike_count=Count('likes', filter=Q(likes__vote=-1))
                ),
                to_attr='annotated_replies'
            )
        ).order_by('-created_date')

        return Article.objects.annotate(
            like_count=Count('likes', filter=Q(likes__vote=1)),
            dislike_count=Count('likes', filter=Q(likes__vote=-1))
        ).prefetch_related(
            Prefetch('comments', queryset=comment_qs, to_attr='annotated_comments'),
            'tag',
        )




class CategoryModelViewSet(viewsets.ReadOnlyModelViewSet):

    serializer_class = CategorySerializer
    queryset = Category.objects.all()
    permission_classes = []



class SubCategoryModelViewSet(viewsets.ReadOnlyModelViewSet):

    serializer_class = SubCategorySerializer
    queryset = SubCategory.objects.all()
    permission_classes = []



class TagModelViewSet(viewsets.ReadOnlyModelViewSet):

    serializer_class = TagSerializer
    queryset = Tag.objects.all()
    permission_classes = []




class CommentModelViewSet(viewsets.ModelViewSet):
    permission_classes = [IsAuthenticatedRemote, IsOwnerOrReadOnly]
    filter_backends = [OrderingFilter]

    ordering_fields = ['created_date', 'likes_count']

    def get_serializer_class(self):

        if self.action == 'create':
            return CommentCreateSerializer
        return CommentListSerializer


    def get_queryset(self):
        qs = Comment.objects.filter(approved=True)
        return qs.annotate(
            like_count=Count('likes', filter=Q(likes__vote=1)),
            dislike_count=Count('likes', filter=Q(likes__vote=-1))
        )
    


    def perform_create(self, serializer):
        user_payload = getattr(self.request, 'remote_user_payload', {})
        if not user_payload:
            raise PermissionDenied("Authentication required. Please provide a valid token.")

        serializer.save(
            username=user_payload.get('username'),
            citizenId=user_payload.get('citizenId'),
            avatar=user_payload.get('avatar'),
            slug_level=user_payload.get('slug_level', 0),
            image_level=user_payload.get('image_level', ''),
        )


    def get_serializer_class(self):

        if self.action == 'create':
            return CommentCreateSerializer

        return CommentListSerializer
    


class ReplyModelViewSet(viewsets.ModelViewSet):

    queryset = Reply.objects.filter(approved=True)
    permission_classes = [IsAuthenticatedRemote, IsOwnerOrReadOnly]
    ordering_fields = ['created_date', 'likes_count']
    filter_backends = [OrderingFilter]


    def get_serializer_class(self):

        if self.action == 'create':
            return ReplyCreateSerializer
        return ReplyListSerializer

    def get_queryset(self):
        return Reply.objects.annotate(like_count=Count('likes', filter=Q(likes__vote=1)), dislike_count=Count('likes', filter=Q(likes__vote=-1)))


    def perform_create(self, serializer):
        user_payload = getattr(self.request, 'remote_user_payload', {})
        if not user_payload:
            raise PermissionDenied("Authentication required. Please provide a valid token.")

        serializer.save(
            username=user_payload.get('username'),
            citizenId=user_payload.get('citizenId'),
            avatar=user_payload.get('avatar'),
            slug_level=user_payload.get('slug_level', 0),
            image_level=user_payload.get('image_level', ''),
        )





class LikeDislikeModelViewSet(viewsets.ModelViewSet):

    serializer_class = LikeDislikeSerializer
    queryset = LikeDislike.objects.all()
    permission_classes = [IsAuthenticatedRemote]

    def perform_create(self, serializer):
        user_payload = getattr(self.request, 'remote_user_payload', {})
        if not user_payload:
            raise PermissionDenied("Authentication required. Please provide a valid token.")
        try:
            serializer.save(
                username=user_payload.get('username'),
                citizenId=user_payload.get('citizenId'),
            )
        except IntegrityError:
            raise ValidationError({"detail": "you liked before"})