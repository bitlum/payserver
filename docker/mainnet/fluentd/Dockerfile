FROM fluent/fluentd:stable

# Install required fluentd plugins
RUN gem install fluent-plugin-concat \
                fluent-plugin-rewrite-tag-filter \
                fluent-plugin-elasticsearch \
                fluent-plugin-prometheus --no-rdoc --no-ri

# Copy fluentd configuration file to place where fluentd expect to have it
COPY ./fluent.conf /fluentd/etc/fluent.conf

# In order to add verboce tracing to fluentd daemon you need to rewrite
# docker CMD directive with and add "-vv" flag.