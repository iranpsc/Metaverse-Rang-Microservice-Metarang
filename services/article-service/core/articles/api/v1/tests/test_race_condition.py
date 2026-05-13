import pytest
import threading
from django.db import connection
from django.db import transaction
from ....models import Article
from .factories import ArticleFactory

pytestmark = pytest.mark.django_db

class TestRaceConditionSlug:
    def test_concurrent_article_creation(self):
        """تست ایجاد همزمان چند مقاله با slug خودکار - فقط برای دیتابیس‌های غیر SQLite"""
        if 'sqlite' in connection.vendor:
            pytest.skip("SQLite does not support high concurrency well")
        
        articles = []
        errors = []

        def create():
            try:
                with transaction.atomic():
                    article = ArticleFactory.create(slug=None)
                    articles.append(article)
            except Exception as e:
                errors.append(e)

        threads = [threading.Thread(target=create) for _ in range(5)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        assert len(articles) == 5, f"Only {len(articles)} created, errors: {errors}"
        slugs = [a.slug for a in articles]
        assert len(slugs) == len(set(slugs)), "Duplicate slugs found"