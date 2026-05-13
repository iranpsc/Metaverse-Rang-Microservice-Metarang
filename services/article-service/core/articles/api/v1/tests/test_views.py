import pytest
import json
from rest_framework import status
from django.urls import reverse
from ....models import Article, Category, SubCategory, Tag, Comment, Reply, LikeDislike
from .factories import (
    ArticleFactory, CategoryFactory, SubCategoryFactory, TagFactory,
    CommentFactory, ReplyFactory, UserFactory
)

pytestmark = pytest.mark.django_db

class TestArticleViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:article-list')
        self.get_detail_url = lambda pk: reverse('article:api-v1:article-detail', args=[pk])
        Article.objects.all().delete()
    
    def test_list_articles_unauthenticated(self, api_client):
        Article.objects.all().delete()
        ArticleFactory.create_batch(3)
        response = api_client.get(self.list_url)
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data['results']) == 3
    
    def test_list_articles_authenticated(self, authenticated_client):
        Article.objects.all().delete()
        ArticleFactory.create_batch(3)
        response = authenticated_client.get(self.list_url)
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data['results']) == 3
    
    def test_create_article_unauthenticated(self, api_client, category, subcategory, tag):
        data = {
            'title': 'New Article',
            'read_time_min': 5,
            'short_description': 'Short',
            'content': 'Content',
            'category': category.id,
            'subcategory': subcategory.id,
            'author': 'Author',
            'identifier': 'id123',
            'email': 'test@example.com',
            'tag': [tag.id],
            'article_image': 'http://example.com/img.jpg',
            'avatar': 'http://example.com/avatar.jpg',
        }
        response = api_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_403_FORBIDDEN

    def test_create_article_authenticated(self, admin_client, test_user, category, subcategory, tag):
        data = {
            'title': 'New Article',
            'slug': 'new-article-123',
            'read_time_min': 5,
            'short_description': 'Short description',
            'content': 'This is the article content',
            'category': category.id,
            'subcategory': subcategory.id,
            'author': 'John Doe',
            'identifier': 'unique-id-123',
            'email': 'john@example.com',
            'tag': [tag.id],
            'article_image': 'http://example.com/test.jpg',
            'avatar': 'http://example.com/avatar.jpg',
        }
        response = admin_client.post(self.list_url, data)
        assert response.status_code == status.HTTP_201_CREATED

    def test_retrieve_article_detail(self, api_client, article):
        url = self.get_detail_url(article.id)
        response = api_client.get(url)
        assert response.status_code == status.HTTP_200_OK
        assert 'content' in response.data

    def test_update_article_authenticated(self, authenticated_client, article, test_user):
        url = self.get_detail_url(article.id)
        data = {'title': 'Updated Title'}
        response = authenticated_client.patch(url, data, format='json')
        assert response.status_code == status.HTTP_200_OK

    def test_delete_article_authenticated(self, authenticated_client, article):
        url = self.get_detail_url(article.id)
        response = authenticated_client.delete(url)
        assert response.status_code == status.HTTP_204_NO_CONTENT

    def test_filter_articles_by_category(self, api_client, category):
        article1 = ArticleFactory(category=category)
        ArticleFactory()
        response = api_client.get(self.list_url, {'category': category.id})
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data['results']) == 1
        assert response.data['results'][0]['id'] == article1.id

    def test_search_articles(self, api_client):
        ArticleFactory(title='Python Programming')
        ArticleFactory(title='Django Framework')
        ArticleFactory(title='JavaScript Basics')
        response = api_client.get(self.list_url, {'search': 'Python'})
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data['results']) == 1

    def test_order_articles_by_created_date(self, api_client):
        article1 = ArticleFactory()
        article2 = ArticleFactory()
        response = api_client.get(self.list_url, {'ordering': '-created_date'})
        assert response.status_code == status.HTTP_200_OK
        assert response.data['results'][0]['id'] == article2.id


    def test_filter_articles_by_multiple_categories(self, api_client, category):
        cat2 = CategoryFactory()
        article1 = ArticleFactory(category=category)
        article2 = ArticleFactory(category=cat2)
        response = api_client.get(self.list_url, {'category__in': f'{category.id},{cat2.id}'})
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data['results']) == 2


    def test_create_article_with_invalid_image(self, authenticated_client, category, subcategory, tag, invalid_image):
        data = {
            'title': 'Invalid',
            'read_time_min': 5,
            'short_description': '...',
            'content': '...',
            'category': category.id,
            'subcategory': subcategory.id,
            'author': 'A',
            'identifier': 'id',
            'email': 'e@e.com',
            'tag': [tag.id],
            'article_image': invalid_image,
        }
        response = authenticated_client.post(self.list_url, data)
        assert response.status_code == status.HTTP_400_BAD_REQUEST



class TestCategoryViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:category-list')
        Article.objects.all().delete()
        Category.objects.all().delete()
    
    def test_list_categories(self, api_client):
        Category.objects.all().delete()
        CategoryFactory.create_batch(3)
        response = api_client.get(self.list_url)
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data) == 3
    

class TestCommentViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:comment-list')
    
    def test_list_comments_unauthenticated(self, api_client, article):
        CommentFactory.create_batch(2, article=article)
        response = api_client.get(self.list_url)
        assert response.status_code == status.HTTP_200_OK
    

    def test_create_comment_unauthenticated(self, api_client, article):
        data = {
            'article': article.id,
            'content': 'Nice article!'
        }
        response = api_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_403_FORBIDDEN

    def test_create_comment_authenticated(self, authenticated_client, article):
        data = {'article': article.id, 'content': 'Great article!'}
        response = authenticated_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_201_CREATED

    def test_only_approved_comments_visible(self, api_client, article):
        CommentFactory(article=article, approved=True)
        CommentFactory(article=article, approved=False)
        response = api_client.get(self.list_url)
        assert response.status_code == status.HTTP_200_OK
        assert len(response.data) == 1

class TestReplyViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:reply-list')
    
    def test_create_reply_authenticated(self, authenticated_client, comment, test_user):
        data = {
            'comment': comment.id,
            'author': test_user.username,
            'content': 'This is a reply'
        }
        response = authenticated_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_201_CREATED
        assert Reply.objects.count() == 1

class TestLikeDislikeViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:like_dislike-list')
    
    def test_like_article_authenticated(self, authenticated_client, article, test_user):
        data = {
            'vote': 1,
            'author': test_user.username,
            'article': article.id
        }
        response = authenticated_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_201_CREATED
        assert LikeDislike.objects.count() == 1

    def test_dislike_article_authenticated(self, authenticated_client, article, test_user):
        data = {
            'vote': -1,
            'author': test_user.username,
            'article': article.id
        }
        response = authenticated_client.post(self.list_url, data, format='json')
        assert response.status_code == status.HTTP_201_CREATED

    def test_like_same_article_twice(self, authenticated_client, article, test_user):
        data = {
            'vote': 1,
            'author': test_user.username,
            'article': article.id
        }
        response1 = authenticated_client.post(self.list_url, data, format='json')
        assert response1.status_code == status.HTTP_201_CREATED
        response2 = authenticated_client.post(self.list_url, data, format='json')
        assert response2.status_code == status.HTTP_400_BAD_REQUEST

    def test_like_count_in_article(self, api_client, article):
        import json
        from django.contrib.auth import get_user_model
        User = get_user_model()
        user1 = User.objects.create_user(username='user1', password='pass')
        user2 = User.objects.create_user(username='user2', password='pass')
        LikeDislike.objects.create(
            username=user1.username,
            citizenId='1111111111',
            vote=1,
            article=article
        )
        LikeDislike.objects.create(
            username=user2.username,
            citizenId='2222222222',
            vote=1,
            article=article
        )
        url = reverse('article:api-v1:article-list')
        response = api_client.get(url)
        assert response.status_code == status.HTTP_200_OK
        article_data = next((item for item in response.data['results'] if item['id'] == article.id), None)
        assert article_data is not None
        assert 'stats' in article_data
        stats = json.loads(article_data['stats'])
        assert stats['likes'] == 2




class TestSubCategoryViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:subcategory-list')
        self.detail_url = lambda pk: reverse('article:api-v1:subcategory-detail', args=[pk])



class TestTagViewSet:
    def setup_method(self):
        self.list_url = reverse('article:api-v1:tag-list')
        self.detail_url = lambda pk: reverse('article:api-v1:tag-detail', args=[pk])


