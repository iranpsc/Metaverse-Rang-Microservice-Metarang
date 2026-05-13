import pytest
from ..serializers import *
from .factories import *
from ....models import *
from django.db.models import Count, Q


pytestmark = pytest.mark.django_db

class TestCategorySerializer:
    """تست‌های سریالایزر Category"""
    
    def test_category_serializer_valid_data(self, sample_image):
        """تست سریالایزر با داده‌های معتبر"""
        data = {
            'name': 'Technology',
            'category_slug': 'technology',
            'category_description': 'Tech news and articles',
            'category_image': sample_image, 

        }
        serializer = CategorySerializer(data=data)
        assert serializer.is_valid(), serializer.errors
    
    def test_category_serializer_missing_required_field(self):
        """تست سریالایزر با فیلد اجباری خالی"""
        data = {
            'name': 'Technology',
            'category_slug': 'technology'
            # category_description missing
        }
        serializer = CategorySerializer(data=data)
        assert not serializer.is_valid()
        assert 'category_description' in serializer.errors

class TestSubCategorySerializer:
    """تست‌های سریالایزر SubCategory"""
    
    def test_subcategory_serializer_valid_data(self):
        """تست سریالایزر با داده‌های معتبر"""
        category = CategoryFactory()
        data = {
            'category': category.id,
            'name': 'Gaming'
        }
        serializer = SubCategorySerializer(data=data)
        assert serializer.is_valid()
    
    def test_subcategory_serializer_invalid_category(self):
        """تست سریالایزر با دسته‌بندی نامعتبر"""
        data = {
            'category': 99999,  # id ناموجود
            'name': 'Gaming'
        }
        serializer = SubCategorySerializer(data=data)
        assert not serializer.is_valid()

class TestTagSerializer:
    """تست‌های سریالایزر Tag"""
    
    def test_tag_serializer_valid_data(self):
        """تست سریالایزر با داده‌های معتبر"""
        data = {
            'label': 'Python',
            'slug': 'python'
        }
        serializer = TagSerializer(data=data)
        assert serializer.is_valid()
    
    def test_tag_serializer_duplicate_slug(self):
        """تست سریالایزر با slug تکراری"""
        TagFactory(slug='python')
        data = {
            'label': 'Python Django',
            'slug': 'python'
        }
        serializer = TagSerializer(data=data)
        assert not serializer.is_valid()

class TestArticleListSerializer:

    def test_article_list_serializer_fields(self, rf):
        """تست فیلدهای سریالایزر لیست مقاله"""
        article = ArticleFactory()
        
        request = rf.get('/article/api/article/')
        
        serializer = ArticleListSerializer(article, context={'request': request})
        data = serializer.data
        
        assert 'id' in data
        assert 'title' in data
        assert 'slug' in data



    def test_article_list_serializer_computed_fields(self, rf):
        """تست فیلدهای محاسباتی سریالایزر"""
        article = ArticleFactory(content="A" * 200)  # بیشتر از 150 کاراکتر
        request = rf.get('/article/api/article/')
        serializer = ArticleListSerializer(article, context={'request': request})
        data = serializer.data

        assert data['snippet'] is not None
        # با منطق 150 کاراکتری مدل
        assert len(data['snippet']) == 154  # 150 + " ..." = 154
        assert data['snippet'].endswith(" ...")



class TestArticleCreateSerializer:
    """تست‌های سریالایزر ایجاد مقاله"""
    
    def test_article_create_serializer_valid_data(self, category, subcategory, tag, sample_image):
        """تست سریالایزر ایجاد با داده‌های معتبر"""
        data = {
            'title': 'New Article',
            'slug': 'new-article',
            'read_time_min': 5,
            'short_description': 'Short desc',
            'content': 'Content here...',
            'category': category.id,
            'subcategory': subcategory.id,
            'author': 'John Doe',
            'identifier': 'test-id-123',
            'email': 'john@example.com',
            'tag': [tag.id],
            'article_image': 'http://example.com/test.jpg',
            'avatar': 'http://example.com/avatar.jpg',
        }
        serializer = ArticleCreateSerializer(data=data)
        assert serializer.is_valid(), serializer.errors
    
    def test_article_create_serializer_missing_required_fields(self):
        """تست سریالایزر ایجاد با فیلدهای اجباری خالی"""
        data = {}  # داده خالی
        serializer = ArticleCreateSerializer(data=data)
        assert not serializer.is_valid()
        assert 'title' in serializer.errors

class TestCommentListSerializer:
    def test_comment_list_serializer_fields(self, article):
        comment = CommentFactory(article=article)
        comment_annotated = Comment.objects.filter(id=comment.id).annotate(
            like_count=Count('likes', filter=Q(likes__vote=1)),
            dislike_count=Count('likes', filter=Q(likes__vote=-1))
        ).first()
        serializer = CommentListSerializer(comment_annotated)
        data = serializer.data
        assert 'id' in data
        assert 'content' in data
        assert 'username' in data
        assert 'replies' in data
        assert 'likes' in data
        assert 'dislikes' in data


class TestCommentCreateSerializer:
    def test_comment_create_serializer_valid_data(self, article):
        data = {'article': article.id, 'content': 'Test comment'}
        serializer = CommentCreateSerializer(data=data)
        assert serializer.is_valid()



class TestReplyListSerializer:
    
    def test_reply_list_serializer_fields(self, comment):
        reply = ReplyFactory(comment=comment)
        reply_annotated = Reply.objects.filter(id=reply.id).annotate(
            like_count=Count('likes', filter=Q(likes__vote=1)),
            dislike_count=Count('likes', filter=Q(likes__vote=-1))
        ).first()
        serializer = ReplyListSerializer(reply_annotated)
        data = serializer.data
        assert 'id' in data
        assert 'content' in data
        assert 'username' in data



class TestLikeDislikeSerializer:
    """تست‌های سریالایزر LikeDislike"""
    
    def test_like_serializer_valid_data(self, article):
        data = {'vote': 1, 'article': article.id}
        serializer = LikeDislikeSerializer(data=data)
        assert serializer.is_valid()
    
    def test_dislike_serializer_valid_data(self, comment):
        """تست سریالایزر دیس‌لایک با داده معتبر"""
        data = {
            'vote': -1,
            'author': 'Test User',
            'comment': comment.id
        }
        serializer = LikeDislikeSerializer(data=data)
        assert serializer.is_valid()