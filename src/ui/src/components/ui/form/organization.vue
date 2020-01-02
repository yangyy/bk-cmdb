<template>
    <div class="cmdb-organization"
        :class="{
            'is-focus': focus,
            'is-disabled': disabled,
            'is-readonly': readonly,
            'is-unselected': isUnselected
        }"
        :data-placeholder="placeholder">
        <i class="select-angle bk-icon icon-angle-down"></i>
        <bk-popover class="select-dropdown"
            ref="selectDropdown"
            trigger="click"
            placement="bottom-start"
            theme="light select-dropdown"
            animation="slide-toggle"
            :z-index="zIndex"
            :arrow="false"
            :offset="-1"
            :distance="12"
            :on-show="handleDropdownShow"
            :on-hide="handleDropdownHide">
            <div class="select-name"
                :title="displayName">
                {{displayName}}
            </div>
            <div slot="content" :style="{ width: popoverWidth + 'px' }" class="select-dropdown-content">
                <div class="search-bar">
                    <bk-input
                        placeholder="搜索"
                        ext-cls="search-input"
                        right-icon="bk-icon icon-search"
                        clearable
                        @input="handleSearch">
                    </bk-input>
                </div>
                <bk-big-tree class="org-tree"
                    ref="tree"
                    v-bkloading="{
                        isLoading: $loading([searchRequestId])
                    }"
                    :show-checkbox="true"
                    :check-on-click="false"
                    :check-strictly="false"
                    :lazy-method="lazyMethod"
                    :lazy-disabled="lazyDisabled"
                    :filter-method="filterMethod"
                    :default-expanded-nodes="defaultCheckedNodes"
                    @check-change="handleCheckChange">
                    <div class="tree-node" slot-scope="{ node, data: nodeData }"
                        :class="{ 'is-selected': node.selected }"
                        :title="nodeData.name">
                        <div class="node-name">{{nodeData.name}}</div>
                    </div>
                </bk-big-tree>
            </div>
        </bk-popover>
    </div>
</template>

<script>
    export default {
        name: 'cmdb-form-organization',
        props: {
            value: {
                type: [Array, String],
                default: []
            },
            disabled: {
                type: Boolean,
                default: false
            },
            readonly: Boolean,
            placeholder: {
                type: String,
                default: '请选择'
            },
            zIndex: {
                type: Number,
                default: 2500
            }
        },
        data () {
            return {
                focus: false,
                checked: this.value,
                defaultCheckedNodes: this.value || [],
                displayName: '',
                popoverWidth: 0,
                searchRequestId: Symbol('orgSearch')
            }
        },
        computed: {
            isUnselected () {
                return this.displayName === ''
            }
        },
        watch: {
            value (value) {
                this.checked = value
            },
            checked (checked) {
                this.$emit('input', checked)
                this.$emit('on-checked', checked)
            }
        },
        created () {
            this.getData()
        },
        methods: {
            async getData (parentId) {
                try {
                    const params = {
                        lookup_field: 'level',
                        exact_lookups: 0
                    }
                    if (parentId) {
                        params.lookup_field = 'parent'
                        params.exact_lookups = parentId
                    }
                    const res = await this.$store.dispatch('organization/getDepartment', { params })
                    const data = res.results || []

                    if (!parentId) {
                        this.setDefaultData(data)
                    } else {
                        return { data }
                    }
                } catch (e) {
                    console.error(e)
                }
            },
            setDefaultData (data) {
                this.$refs.tree.setData(data)
            },
            lazyMethod (node) {
                return this.getData(node.id)
            },
            lazyDisabled (node) {
                return !node.data.has_children
            },
            filterMethod (searchResult, node) {
                console.log(searchResult, node, 'searchResult, node')
                return searchResult.some(item => item.id === node.id)
            },
            async handleSearch (value) {
                try {
                    if (value.length) {
                        const res = await this.$store.dispatch('organization/getDepartment', {
                            params: {
                                lookup_field: 'name',
                                fuzzy_lookups: value
                            },
                            requestId: this.searchRequestId
                        })
                        const data = res.results || []
                        this.$refs.tree.filter(data)
                    } else {
                        this.$refs.tree.filter()
                    }
                } catch (e) {
                    console.error(e)
                }
            },
            handleCheckChange (ids) {
                this.displayName = ids.map(id => this.$refs.tree.getNodeById(id)).map(node => node.data.full_name).join(' ; ')
                this.checked = ids
            },
            handleDropdownShow () {
                this.popoverWidth = this.$el.offsetWidth
                this.focus = true
            },
            handleDropdownHide () {
                this.focus = false
            }
        }
    }
</script>

<style lang="scss">
    .tippy-tooltip {
        &.select-dropdown-theme {
            padding: 0;
            box-shadow: 0 3px 9px 0 rgba(0, 0, 0, .1);
        }
    }

    .search-input {
        .bk-form-input {
            &:focus {
                border-color: #c4c6cc !important;
            }
        }
    }
</style>
<style lang="scss" scoped>
    .cmdb-organization {
        position: relative;
        border: 1px solid #c4c6cc;
        border-radius: 2px;
        line-height: 30px;
        color: #63656e;
        cursor: pointer;
        font-size: 12px;

        &.is-focus {
            border-color: #3a84ff;
            box-shadow:0px 0px 4px rgba(58, 132, 255, 0.4);
            .select-angle {
                transform: rotate(-180deg);
            }
        }
        &.is-disabled {
            background-color: #fafafa;
            color: #c4c6cc;
            cursor: not-allowed;
        }
        &.is-readonly,
        &.is-loading {
            background-color: #fafafa;
            cursor: default;
        }

        &.is-unselected::before {
            position: absolute;
            height: 100%;
            content: attr(data-placeholder);
            left: 10px;
            top: 0;
            color: #c3cdd7;
            pointer-events: none;
        }

        .select-angle {
            position: absolute;
            right: 12px;
            top: 10px;
            font-size: 12px;
            transition: transform .3s cubic-bezier(0.4, 0, 0.2, 1);
            pointer-events: none;
        }

        .select-dropdown {
            display: block;

            .select-name {
                height: 30px;
                padding: 0 36px 0 10px;
                @include ellipsis;
            }
        }
    }

    .select-dropdown-content {
        border: 1px solid #dcdee5;
        border-radius: 2px;
        line-height: 32px;
        background: #fff;
        color: #63656e;
        overflow: hidden;

        .search-bar {
            padding: 10px;
        }

        .org-tree {
            height: 220px;

            .tree-node {
                .node-name {
                    @include ellipsis;
                }
            }
        }
    }

    /deep/.bk-tooltip {
        > .bk-tooltip-ref {
            display: block;
        }
    }
</style>
