
from rest_framework import serializers
from ...models import Article, Category, SubCategory, Tag, Comment, Reply, LikeDislike
import json



class ReplyListSerializer(serializers.ModelSerializer):

    like_count = serializers.IntegerField(read_only=True)
    dislike_count = serializers.IntegerField(read_only=True)
    

    class Meta:
        model = Reply
        fields = ['id', 'comment', 'username', 'citizenId', 'avatar', 'slug_level', 'image_level', 'content',  'approved', 'created_date', 'like_count', 'dislike_count']
        read_only_fields = ['approved', 'created_date']


    def to_representation(self, instance):
        request = self.context.get('request')
        rep = super().to_representation(instance)

        rep['likes'] = rep.pop('like_count')
        rep['dislikes'] = rep.pop('dislike_count')
        return rep




class ReplyCreateSerializer(serializers.ModelSerializer):

    class Meta:
        model = Reply
        fields = ['id', 'comment', 'content']
        read_only_fields = ['username', 'citizenId', 'avatar', 'slug_level', 'image_level', 'approved']





class CommentListSerializer(serializers.ModelSerializer):
    replies = serializers.SerializerMethodField()

    like_count = serializers.IntegerField(read_only=True)
    dislike_count = serializers.IntegerField(read_only=True)

    class Meta:
        model = Comment
        fields = ['id', 'article', 'username', 'citizenId', 'avatar', 'slug_level', 'image_level', 'content', 'approved', 'created_date', 'like_count', 'dislike_count', 'replies']


    def get_replies(self, obj):
        annotated_replies = getattr(obj, 'annotated_replies', [])
        return ReplyListSerializer(annotated_replies, many=True, context=self.context).data
    

    def to_representation(self, instance):
        request = self.context.get('request')
        rep = super().to_representation(instance)

        rep['likes'] = rep.pop('like_count')
        rep['dislikes'] = rep.pop('dislike_count')
        return rep
    


class CommentCreateSerializer(serializers.ModelSerializer):

    class Meta:
        model = Comment
        fields = ['id', 'article', 'content']
        read_only_fields = ['username', 'citizenId', 'avatar', 'slug_level', 'image_level', 'approved']




class ArticleListSerializer(serializers.ModelSerializer):

    snippet = serializers.ReadOnlyField(source='get_snippet')

    relative_url = serializers.URLField(source='get_absolute_api_url', read_only=True)
    absolute_url = serializers.SerializerMethodField()

    comments = serializers.SerializerMethodField()

    like_count = serializers.IntegerField(read_only=True)
    dislike_count = serializers.IntegerField(read_only=True)


    class Meta:
        model = Article
        fields = ['id', 'title', 'slug', 'read_time_min', 'article_image', 'short_description', 'snippet', 'content', 'relative_url',
                'absolute_url', 'category', 'subcategory', 'author','identifier', 'avatar', 'field_of_activity', 'biography', 'telegram_id',
                'whatsapp_number', 'email', 'like_count', 'dislike_count', 'created_date', 'comments']






 
    def to_representation(self, instance):
        request = self.context.get('request')
        rep = super().to_representation(instance)

        category_data = CategorySerializer(instance.category, context={'request': request}).data

        rep['category'] = category_data.get('name', '')
        rep['categorySlug'] = category_data.get('categorySlug', '')
        rep['categoryImage'] = category_data.get('categoryImage', '')
        rep['categoryDec'] = category_data.get('categoryDec', '')

        subcat_data = SubCategorySerializer(instance.subcategory, context={'request': request}).data
        rep['subcategory'] = subcat_data.get('name', '')

        author_info = {
            "bio": rep.pop('biography', ''),
            "name": rep.pop('author', ''),
            "field": rep.pop('field_of_activity', ''),
            "avatar": rep.pop('avatar', ''),
            "socials": {
                "email": rep.pop('email', ''),
                "telegram": rep.pop('telegram_id', ''),
                "whatsapp": rep.pop('whatsapp_number', '')
            },
            "citizenId": rep.pop('identifier', '')
        }
        rep['author'] = json.dumps(author_info, ensure_ascii=False)

        stats_info = {
            "likes": rep.pop('like_count', 0),
            "dislikes": rep.pop('dislike_count', 0),
            "comments": len(rep.get('comments', [])), 
            "views": 0
        }
        rep['stats'] = json.dumps(stats_info, ensure_ascii=False)

        tags_data = TagSerializer(instance.tag.all(), context={'request': request}, many=True).data
        rep['tags'] = json.dumps(tags_data, ensure_ascii=False)

        rep['readingTime'] = rep.pop('read_time_min')
        rep['image'] = rep.pop('article_image')
        rep['description'] = rep.pop('short_description')
        rep['subCategory'] = rep.pop('subcategory')  
        
       
        rep['date'] = rep.pop('created_date')

        if request and hasattr(request, 'parser_context'):
            kwargs = request.parser_context.get('kwargs', {})
            if kwargs.get('pk'):
                rep.pop('snippet', None)
                rep.pop('relative_url', None)
                rep.pop('absolute_url', None)
            else: 
                rep.pop('content', None)
        else:
            rep.pop('content', None)

        return rep

    

    def get_absolute_url(self, obj):
        request = self.context.get('request')
        if request is None:
            return None
        return request.build_absolute_uri(f'/article/api/article/{obj.pk}/')
    

    def get_comments(self, obj):
        annotated_comments = getattr(obj, 'annotated_comments', [])
        return CommentListSerializer(annotated_comments, many=True, context=self.context).data



class ArticleCreateSerializer(serializers.ModelSerializer):
    article_image = serializers.URLField(required=True, write_only=True)
    avatar = serializers.URLField(required=True, write_only=True)


    class Meta:
        model = Article
        fields = ['id', 'title', 'slug', 'read_time_min', 'article_image', 'short_description', 'content','category', 'subcategory', 'author',
                'identifier', 'avatar', 'field_of_activity', 'biography', 'telegram_id',
                'whatsapp_number', 'email', 'tag']
        
        read_only_fields = ['slug']

    def create(self, validated_data):
        tags_data = validated_data.pop('tag', [])
        article = Article.objects.create(**validated_data)
        article.tag.set(tags_data)
        return article







class CategorySerializer(serializers.ModelSerializer):
    class Meta:
        model = Category
        fields = ['id', 'name', 'category_slug', 'category_image', 'category_description']


    def to_representation(self, instance):
        request = self.context.get('request')
        rep = super().to_representation(instance)

        rep['categorySlug'] = rep.pop('category_slug')
        rep['categoryImage'] = rep.pop('category_image')
        rep['categoryDec'] = rep.pop('category_description')
        return rep




class SubCategorySerializer(serializers.ModelSerializer):
    class Meta:
        model = SubCategory
        fields = ['id','category', 'name']



class TagSerializer(serializers.ModelSerializer):
    class Meta:
        model = Tag
        fields = ['id', 'label', 'slug']




class LikeDislikeSerializer(serializers.ModelSerializer):


    class Meta: 
        model = LikeDislike
        fields = ['vote', 'article', 'comment', 'reply', 'created_date']
        read_only_fields = ['username', 'citizenId']

