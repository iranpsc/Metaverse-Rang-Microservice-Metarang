import pytest
from django.core.exceptions import ValidationError
from django.db import IntegrityError, transaction
from django.contrib.auth import get_user_model
from ....models import Article, Category, SubCategory, Tag, Comment, Reply, LikeDislike
from .factories import (
    CategoryFactory, SubCategoryFactory, TagFactory, ArticleFactory,
    CommentFactory, ReplyFactory, UserFactory
)
import django.db
from django.db import connection
import threading



User = get_user_model()

pytestmark = pytest.mark.django_db

class TestCategoryModel:
    """تست‌های مدل Category"""
    
    def test_create_category(self):
        """تست ایجاد دسته‌بندی"""
        category = CategoryFactory()
        assert category.pk is not None
        assert isinstance(category.name, str)
    
    def test_category_str_method(self):
        """تست متد __str__"""
        category = CategoryFactory(name="Technology")
        assert str(category) == "Technology"



    def test_category_unique_name(self):
        if 'sqlite' in connection.vendor:
            pytest.skip("SQLite handles unique constraints differently")
        
        CategoryFactory(name="Unique")
        with pytest.raises(IntegrityError):
            CategoryFactory(name="Unique")



    


            
class TestSubCategoryModel:
    """تست‌های مدل SubCategory"""
    
    def test_create_subcategory(self):
        """تست ایجاد زیردسته‌بندی"""
        subcategory = SubCategoryFactory()
        assert subcategory.pk is not None
        assert subcategory.category is not None
    
    def test_subcategory_str_method(self):
        """تست متد __str__"""
        subcategory = SubCategoryFactory(name="Gaming")
        assert str(subcategory) == "Gaming"
    
    def test_subcategory_relationship_with_category(self):
        """تست رابطه زیردسته‌بندی با دسته‌بندی"""
        category = CategoryFactory()
        subcategory = SubCategoryFactory(category=category)
        assert subcategory.category == category
        assert subcategory in category.subcategories.all()

class TestTagModel:
    """تست‌های مدل Tag"""
    
    def test_create_tag(self):
        """تست ایجاد تگ"""
        tag = TagFactory()
        assert tag.pk is not None
        assert tag.label is not None
        assert tag.slug is not None
    
    def test_tag_str_method(self):
        """تست متد __str__"""
        tag = TagFactory(label="Python")
        assert str(tag) == "Python"
    
    def test_tag_unique_slug(self):
        """تست یکتا بودن slug تگ"""
        TagFactory(slug="unique-slug")
        with pytest.raises(IntegrityError):
            TagFactory(slug="unique-slug")

class TestArticleModel:
    """تست‌های مدل Article"""
    
    def test_create_article(self):
        """تست ایجاد مقاله"""
        article = ArticleFactory()
        assert article.pk is not None
        assert article.title is not None
        assert article.slug is not None
    
    def test_article_str_method(self):
        """تست متد __str__"""
        article = ArticleFactory(title="My Article")
        assert str(article) == "My Article"
    
    def test_article_slug_validation(self):
        """تست اعتبارسنجی slug - فقط کاراکترهای مجاز انگلیسی"""
        article = ArticleFactory.build()
        article.slug = "invalid slug with space"
        with pytest.raises(ValidationError):
            article.full_clean()
    
    def test_article_slug_auto_generation_on_create(self):
        """تست تولید خودکار slug هنگام ایجاد"""
        article1 = ArticleFactory()
        article2 = ArticleFactory()
        assert article1.slug.startswith('bl-')
        assert article2.slug.startswith('bl-')
        assert article1.slug != article2.slug



    def test_get_snippet_method(self):
        """تست متد get_snippet"""
        article = ArticleFactory(content="A" * 200)
        snippet = article.get_snippet()
        # " ..." = 4 کاراکتر (space + 3 dot)
        assert len(snippet) == 154
        assert snippet.endswith(" ...")


    
    def test_get_absolute_api_url(self):
        """تست متد get_absolute_api_url"""
        article = ArticleFactory()
        url = article.get_absolute_api_url()
        assert f"/api/v1/article/{article.pk}" in url or f"article/{article.pk}" in url
    
    def test_article_relationships(self):
        """تست روابط مقاله با مدل‌های دیگر"""
        category = CategoryFactory()
        subcategory = SubCategoryFactory(category=category)
        tag1 = TagFactory()
        tag2 = TagFactory()
        
        article = ArticleFactory(category=category, subcategory=subcategory)
        article.tag.add(tag1, tag2)
        
        assert article.category == category
        assert article.subcategory == subcategory
        assert tag1 in article.tag.all()
        assert tag2 in article.tag.all()

class TestCommentModel:
    """تست‌های مدل Comment"""
    
    def test_create_comment(self):
        """تست ایجاد کامنت"""
        comment = CommentFactory()
        assert comment.pk is not None
        assert comment.content is not None
        assert comment.approved is True
    
    def test_comment_str_method(self):
        """تست متد __str__"""
        article = ArticleFactory(title="Test Article")
        comment = CommentFactory(article=article, username="John Doe")
        assert "John Doe" in str(comment)
        assert "Test Article" in str(comment)
    
    def test_comment_ordering(self):
        """تست مرتب‌سازی کامنت‌ها بر اساس تاریخ ایجاد (نزولی)"""
        article = ArticleFactory()
        comment1 = CommentFactory(article=article)
        comment2 = CommentFactory(article=article)
        comment3 = CommentFactory(article=article)
        
        comments = Comment.objects.filter(article=article)
        assert comments[0].created_date >= comments[1].created_date
    
    def test_comment_relationship_with_article(self):
        """تست رابطه کامنت با مقاله"""
        article = ArticleFactory()
        comment = CommentFactory(article=article)
        assert comment.article == article
        assert comment in article.comments.all()

class TestReplyModel:
    """تست‌های مدل Reply"""
    
    def test_create_reply(self):
        """تست ایجاد ریپلای"""
        reply = ReplyFactory()
        assert reply.pk is not None
        assert reply.content is not None
        assert reply.approved is True
    
    def test_reply_str_method(self):
        """تست متد __str__"""
        comment = CommentFactory(username="Jane Doe")
        reply = ReplyFactory(comment=comment, username="John Smith")
        assert "John Smith" in str(reply)
        assert "Jane Doe" in str(reply)
    
    def test_reply_relationship_with_comment(self):
        """تست رابطه ریپلای با کامنت"""
        comment = CommentFactory()
        reply = ReplyFactory(comment=comment)
        assert reply.comment == comment
        assert reply in comment.replies.all()

class TestLikeDislikeModel:
    """تست‌های مدل LikeDislike"""
    
    def test_create_like_for_article(self, article):
        like = LikeDislike.objects.create(
            vote=LikeDislike.LIKE,
            username="testuser",
            citizenId="123456",
            article=article
        )
        assert like.pk is not None
        
    
    def test_create_dislike_for_comment(self):
        """تست ایجاد دیس‌لایک برای کامنت"""
        comment = CommentFactory()
        dislike = LikeDislike.objects.create(
            vote=LikeDislike.DISLIKE,
           username="testuser",
            citizenId="123456",
            comment=comment,
        )
        assert dislike.vote == LikeDislike.DISLIKE
    
    def test_unique_constraint_user_article(self):
        """تست عدم امکان لایک/دیس‌لایک تکراری برای یک مقاله توسط یک کاربر"""
        article = ArticleFactory()
        LikeDislike.objects.create(
            vote=LikeDislike.LIKE,
            username="testuser",
            citizenId="123456",
            article=article
        )
        with pytest.raises(IntegrityError):
            LikeDislike.objects.create(
            vote=LikeDislike.DISLIKE,
            username="testuser",
            citizenId="123456",
            article=article
            )
    
    def test_exactly_one_target_constraint(self):
        """تست اینکه دقیقاً یک هدف (مقاله/کامنت/ریپلای) باید مشخص شود"""
        with pytest.raises(ValidationError):
            like = LikeDislike(
            vote=LikeDislike.LIKE,
            username="testuser",
            citizenId="123456",
            article=None,
            comment=None,
            reply=None
            )
            like.clean()
    
    def test_multiple_targets_not_allowed(self):
        """تست اینکه نمی‌توان همزمان دو هدف مختلف داشت"""
        article = ArticleFactory()
        comment = CommentFactory()
        
        with pytest.raises(ValidationError):
            like = LikeDislike(
            vote=LikeDislike.LIKE,
            username="testuser",
            citizenId="123456",
            article=article,
            comment=comment,
            reply=None
            )
            like.clean()
    
    def test_like_str_method(self):
        """تست متد __str__ برای لایک"""
        article = ArticleFactory(title="Cool Article")
        like = LikeDislike.objects.create(
        vote=LikeDislike.LIKE,
        username="testuser",
        citizenId="123456",
        article=article
        )
        assert "Like" in str(like)
        assert "Cool Article" in str(like)
    
    def test_dislike_str_method(self):
        """تست متد __str__ برای دیس‌لایک"""
        comment = CommentFactory()
        dislike = LikeDislike.objects.create(
        vote=LikeDislike.DISLIKE,
        username="testuser",
        citizenId="123456",
        comment=comment
        )
        assert "Dislike" in str(dislike)





class TestArticleModelExtended:
    def test_slug_race_condition(self):
        """تست همزمانی ایجاد مقاله با slug خودکار (شبیه‌سازی با threading)"""
        articles = []
        errors = []

        def create_article():
            try:
                with transaction.atomic():
                    article = ArticleFactory(slug=None)  
                    articles.append(article)
            except Exception as e:
                errors.append(e)

        threads = []
        for _ in range(5):
            t = threading.Thread(target=create_article)
            threads.append(t)
            t.start()
        for t in threads:
            t.join()
        assert len(articles) + len(errors) >= 5
        slugs = [a.slug for a in articles]
        assert len(slugs) == len(set(slugs)), "Duplicate slugs found"

    def test_slug_auto_generation_on_create(self):
        article1 = ArticleFactory()
        article2 = ArticleFactory()
        assert article1.slug.startswith('bl-')
        assert article2.slug.startswith('bl-')
        assert article1.slug != article2.slug
        num1 = int(article1.slug.split('-')[1])
        num2 = int(article2.slug.split('-')[1])
        assert abs(num1 - num2) == 1

class TestLikeDislikeModelExtended:
    def test_unique_constraint_user_comment(self, comment):
        user_data = {'username': 'testuser', 'citizenId': '123'}
        LikeDislike.objects.create(vote=1, username='testuser', citizenId='123', comment=comment)
        with pytest.raises(IntegrityError):
            LikeDislike.objects.create(vote=-1, username='testuser', citizenId='123', comment=comment)

    def test_unique_constraint_user_reply(self, reply):
        LikeDislike.objects.create(vote=1, username='testuser', citizenId='123', reply=reply)
        with pytest.raises(IntegrityError):
            LikeDislike.objects.create(vote=-1, username='testuser', citizenId='123', reply=reply)

    def test_clean_method_multiple_targets(self, article, comment):
        like = LikeDislike(vote=1, username='u', citizenId='1', article=article, comment=comment)
        with pytest.raises(ValidationError):
            like.clean()