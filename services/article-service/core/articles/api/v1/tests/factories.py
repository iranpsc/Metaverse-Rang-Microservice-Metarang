import factory
from factory.django import DjangoModelFactory
from django.contrib.auth import get_user_model
from faker import Faker
from ....models import Category, SubCategory, Tag, Article, Comment, Reply, LikeDislike

fake = Faker()
User = get_user_model()

class UserFactory(DjangoModelFactory):
    class Meta:
        model = User
        django_get_or_create = ('username',)
        skip_postgeneration_save = True 

    username = factory.Sequence(lambda n: f'user_{n}')
    email = factory.Sequence(lambda n: f'user{n}@example.com')
    password = factory.PostGenerationMethodCall('set_password', 'testpass123')
    first_name = factory.Faker('first_name')
    last_name = factory.Faker('last_name')

class CategoryFactory(DjangoModelFactory):
    class Meta:
        model = Category
        django_get_or_create = ('name',)

    name = factory.Sequence(lambda n: f'Category {n}')
    category_slug = factory.Sequence(lambda n: f'category-slug-{n}')
    category_image = factory.django.ImageField()
    category_description = factory.Faker('text', max_nb_chars=200)

class SubCategoryFactory(DjangoModelFactory):
    class Meta:
        model = SubCategory
        django_get_or_create = ('name',)

    category = factory.SubFactory(CategoryFactory)
    name = factory.Sequence(lambda n: f'SubCategory {n}')

class TagFactory(DjangoModelFactory):
    class Meta:
        model = Tag
        django_get_or_create = ('label',)

    label = factory.Sequence(lambda n: f'Tag {n}')
    slug = factory.Sequence(lambda n: f'tag-slug-{n}')

class ArticleFactory(DjangoModelFactory):
    class Meta:
        model = Article

    title = factory.Sequence(lambda n: f'Article Title {n}')
    slug = factory.Sequence(lambda n: f'article-slug-{n}')
    read_time_min = factory.Faker('random_int', min=1, max=60)
    article_image = factory.django.ImageField()
    short_description = factory.Faker('text', max_nb_chars=200)
    content = factory.Faker('text', max_nb_chars=5000)
    category = factory.SubFactory(CategoryFactory)
    subcategory = factory.SubFactory(SubCategoryFactory)
    author = factory.Faker('name')
    identifier = factory.Faker('uuid4')
    avatar = factory.django.ImageField()
    field_of_activity = factory.Faker('job')
    biography = factory.Faker('text', max_nb_chars=200)
    telegram_id = factory.Faker('user_name')
    whatsapp_number = factory.Faker('phone_number')
    email = factory.Faker('email')

class CommentFactory(DjangoModelFactory):
    class Meta:
        model = Comment

    article = factory.SubFactory(ArticleFactory)
    username = factory.Faker('user_name')
    citizenId = factory.Faker('ssn')
    avatar = factory.Faker('url')
    slug_level = factory.Faker('random_int', min=1, max=12)
    image_level = factory.Faker('url')
    content = factory.Faker('text', max_nb_chars=500)
    approved = True

class ReplyFactory(DjangoModelFactory):
    class Meta:
        model = Reply

    comment = factory.SubFactory(CommentFactory)
    username = factory.Faker('user_name')
    citizenId = factory.Faker('ssn')
    avatar = factory.Faker('url')
    slug_level = factory.Faker('random_int', min=1, max=10)
    image_level = factory.Faker('url')
    content = factory.Faker('text', max_nb_chars=300)
    approved = True

class LikeDislikeFactory(DjangoModelFactory):
    class Meta:
        model = LikeDislike

    username = factory.Faker('user_name')
    citizenId = factory.Faker('ssn')
    vote = factory.Faker('random_element', elements=[1, -1])
    article = None
    comment = None
    reply = None